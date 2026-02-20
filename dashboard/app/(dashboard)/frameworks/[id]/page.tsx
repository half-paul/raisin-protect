'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import {
  ArrowLeft,
  Search,
  ChevronLeft,
  ChevronRight,
  Shield,
  AlertCircle,
  CheckCircle2,
  MinusCircle,
  FileText,
  Target,
} from 'lucide-react';
import {
  OrgFramework,
  CoverageData,
  CoverageRequirement,
  Requirement,
  ScopingDecision,
  listOrgFrameworks,
  getOrgFrameworkCoverage,
  listRequirements,
  listScoping,
  setRequirementScope,
  resetRequirementScope,
} from '@/lib/api';

export default function FrameworkDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { hasRole } = useAuth();
  const canManage = hasRole('ciso', 'compliance_manager');
  const orgFwId = params.id as string;

  const [orgFramework, setOrgFramework] = useState<OrgFramework | null>(null);
  const [coverage, setCoverage] = useState<CoverageData | null>(null);
  const [requirements, setRequirements] = useState<Requirement[]>([]);
  const [scopingDecisions, setScopingDecisions] = useState<ScopingDecision[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('coverage');

  // Coverage filters
  const [coverageFilter, setCoverageFilter] = useState<string>('');
  const [coveragePage, setCoveragePage] = useState(1);
  const [coverageSearch, setCoverageSearch] = useState('');

  // Scoping modal
  const [showScopeModal, setShowScopeModal] = useState(false);
  const [scopeTarget, setScopeTarget] = useState<{ id: string; identifier: string; title: string } | null>(null);
  const [scopeForm, setScopeForm] = useState({ in_scope: false, justification: '' });
  const [scopeLoading, setScopeLoading] = useState(false);
  const [scopeError, setScopeError] = useState('');

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      // Fetch org frameworks to find ours
      const orgFwRes = await listOrgFrameworks();
      const found = orgFwRes.data?.find((of) => of.id === orgFwId);
      if (!found) {
        router.push('/frameworks');
        return;
      }
      setOrgFramework(found);

      // Fetch coverage
      const covRes = await getOrgFrameworkCoverage(orgFwId, {
        status: coverageFilter || undefined,
        page: String(coveragePage),
        per_page: '50',
      } as Record<string, string>);
      setCoverage(covRes.data);

      // Fetch requirements tree
      const reqRes = await listRequirements(
        found.framework.id,
        found.active_version.id,
        { format: 'flat', per_page: '200' }
      );
      setRequirements(reqRes.data || []);

      // Fetch scoping decisions
      const scopeRes = await listScoping(orgFwId);
      setScopingDecisions(scopeRes.data || []);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, [orgFwId, coverageFilter, coveragePage, router]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  async function handleSetScope() {
    if (!scopeTarget) return;
    setScopeError('');
    setScopeLoading(true);
    try {
      await setRequirementScope(orgFwId, scopeTarget.id, {
        in_scope: scopeForm.in_scope,
        justification: scopeForm.justification || undefined,
      });
      setShowScopeModal(false);
      fetchAll();
    } catch (err) {
      setScopeError(err instanceof Error ? err.message : 'Failed to set scope');
    } finally {
      setScopeLoading(false);
    }
  }

  async function handleResetScope(requirementId: string) {
    try {
      await resetRequirementScope(orgFwId, requirementId);
      fetchAll();
    } catch {
      // handle
    }
  }

  function openScopeModal(req: { id: string; identifier: string; title: string }, inScope: boolean) {
    setScopeTarget(req);
    setScopeForm({ in_scope: inScope, justification: '' });
    setScopeError('');
    setShowScopeModal(true);
  }

  // Build a set of out-of-scope requirement IDs
  const outOfScopeIds = new Set(
    scopingDecisions.filter((s) => !s.in_scope).map((s) => s.requirement.id)
  );

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center py-24">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  if (!orgFramework) return null;

  return (
    <div className="p-6 space-y-6">
      {/* Back + Header */}
      <div>
        <Button variant="ghost" size="sm" className="mb-2" onClick={() => router.push('/frameworks')}>
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Frameworks
        </Button>
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{orgFramework.framework.name}</h1>
            <p className="text-muted-foreground mt-1">
              {orgFramework.active_version.display_name} &middot;{' '}
              {orgFramework.active_version.total_requirements} requirements
            </p>
          </div>
          <Badge variant={orgFramework.status === 'active' ? 'default' : 'secondary'} className="text-sm">
            {orgFramework.status}
          </Badge>
        </div>
      </div>

      {/* Summary Cards */}
      {coverage && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
          <Card>
            <CardContent className="pt-6 text-center">
              <div className={`text-3xl font-bold ${
                coverage.summary.coverage_pct >= 80
                  ? 'text-green-600 dark:text-green-400'
                  : coverage.summary.coverage_pct >= 50
                  ? 'text-amber-600 dark:text-amber-400'
                  : 'text-red-600 dark:text-red-400'
              }`}>
                {coverage.summary.coverage_pct.toFixed(1)}%
              </div>
              <p className="text-xs text-muted-foreground mt-1">Coverage</p>
              <Progress value={coverage.summary.coverage_pct} className="h-1.5 mt-2" />
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6 text-center">
              <div className="text-3xl font-bold">{coverage.summary.assessable_requirements}</div>
              <p className="text-xs text-muted-foreground mt-1">Assessable Reqs</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6 text-center">
              <div className="text-3xl font-bold">{coverage.summary.in_scope}</div>
              <p className="text-xs text-muted-foreground mt-1">In Scope</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6 text-center">
              <div className="text-3xl font-bold text-green-600 dark:text-green-400">{coverage.summary.covered}</div>
              <p className="text-xs text-muted-foreground mt-1">Covered</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6 text-center">
              <div className="text-3xl font-bold text-red-600 dark:text-red-400">{coverage.summary.gaps}</div>
              <p className="text-xs text-muted-foreground mt-1">Gaps</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="coverage">
            <Target className="h-4 w-4 mr-1" />
            Coverage
          </TabsTrigger>
          <TabsTrigger value="requirements">
            <FileText className="h-4 w-4 mr-1" />
            Requirements
          </TabsTrigger>
          <TabsTrigger value="scoping">
            <MinusCircle className="h-4 w-4 mr-1" />
            Scoping ({scopingDecisions.length})
          </TabsTrigger>
        </TabsList>

        {/* Coverage Tab */}
        <TabsContent value="coverage" className="space-y-4 mt-4">
          <div className="flex items-center gap-4">
            <div className="flex gap-2">
              <Button
                variant={coverageFilter === '' ? 'default' : 'outline'}
                size="sm"
                onClick={() => { setCoverageFilter(''); setCoveragePage(1); }}
              >
                All
              </Button>
              <Button
                variant={coverageFilter === 'covered' ? 'default' : 'outline'}
                size="sm"
                onClick={() => { setCoverageFilter('covered'); setCoveragePage(1); }}
              >
                <CheckCircle2 className="h-3 w-3 mr-1" />
                Covered
              </Button>
              <Button
                variant={coverageFilter === 'gap' ? 'default' : 'outline'}
                size="sm"
                onClick={() => { setCoverageFilter('gap'); setCoveragePage(1); }}
              >
                <AlertCircle className="h-3 w-3 mr-1" />
                Gaps
              </Button>
            </div>
          </div>

          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[120px]">Requirement</TableHead>
                    <TableHead>Title</TableHead>
                    <TableHead className="w-[100px]">Status</TableHead>
                    <TableHead>Controls</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {coverage?.requirements?.map((req) => (
                    <TableRow key={req.id}>
                      <TableCell className="font-mono text-sm font-medium">{req.identifier}</TableCell>
                      <TableCell className="max-w-md">
                        <span className="line-clamp-2 text-sm">{req.title}</span>
                      </TableCell>
                      <TableCell>
                        {req.status === 'covered' ? (
                          <Badge className="bg-green-500/10 text-green-700 dark:text-green-400 hover:bg-green-500/20">
                            <CheckCircle2 className="h-3 w-3 mr-1" />
                            Covered
                          </Badge>
                        ) : (
                          <Badge variant="destructive" className="bg-red-500/10 text-red-700 dark:text-red-400 hover:bg-red-500/20">
                            <AlertCircle className="h-3 w-3 mr-1" />
                            Gap
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {req.controls.length > 0 ? (
                            req.controls.map((ctrl) => (
                              <Link key={ctrl.id} href={`/controls/${ctrl.id}`}>
                                <Badge variant="outline" className="text-xs cursor-pointer hover:bg-accent">
                                  {ctrl.identifier}
                                  {ctrl.strength !== 'primary' && (
                                    <span className="ml-1 opacity-60">({ctrl.strength})</span>
                                  )}
                                </Badge>
                              </Link>
                            ))
                          ) : (
                            <span className="text-xs text-muted-foreground">No controls mapped</span>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                  {(!coverage?.requirements || coverage.requirements.length === 0) && (
                    <TableRow>
                      <TableCell colSpan={4} className="text-center py-8 text-muted-foreground">
                        No requirements found
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Requirements Tab */}
        <TabsContent value="requirements" className="space-y-4 mt-4">
          <div className="relative max-w-sm">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search requirements..."
              className="pl-10"
              value={coverageSearch}
              onChange={(e) => setCoverageSearch(e.target.value)}
            />
          </div>

          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[120px]">ID</TableHead>
                    <TableHead>Title</TableHead>
                    <TableHead className="w-[80px]">Depth</TableHead>
                    <TableHead className="w-[100px]">Assessable</TableHead>
                    <TableHead className="w-[100px]">Scope</TableHead>
                    {canManage && <TableHead className="w-[120px]">Actions</TableHead>}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {requirements
                    .filter(
                      (r) =>
                        !coverageSearch ||
                        r.identifier.toLowerCase().includes(coverageSearch.toLowerCase()) ||
                        r.title.toLowerCase().includes(coverageSearch.toLowerCase())
                    )
                    .map((req) => {
                      const isOutOfScope = outOfScopeIds.has(req.id);
                      return (
                        <TableRow key={req.id} className={isOutOfScope ? 'opacity-50' : ''}>
                          <TableCell className="font-mono text-sm font-medium">
                            <span style={{ paddingLeft: `${req.depth * 16}px` }}>
                              {req.identifier}
                            </span>
                          </TableCell>
                          <TableCell className="max-w-lg">
                            <span className="line-clamp-2 text-sm">{req.title}</span>
                          </TableCell>
                          <TableCell className="text-center text-muted-foreground">{req.depth}</TableCell>
                          <TableCell>
                            {req.is_assessable ? (
                              <Badge variant="outline" className="text-xs">Yes</Badge>
                            ) : (
                              <span className="text-xs text-muted-foreground">Section</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {isOutOfScope ? (
                              <Badge variant="secondary" className="text-xs">Out of scope</Badge>
                            ) : (
                              <Badge className="text-xs bg-green-500/10 text-green-700 dark:text-green-400">In scope</Badge>
                            )}
                          </TableCell>
                          {canManage && (
                            <TableCell>
                              {req.is_assessable && (
                                <div className="flex gap-1">
                                  {isOutOfScope ? (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="text-xs h-7"
                                      onClick={() => handleResetScope(req.id)}
                                    >
                                      Include
                                    </Button>
                                  ) : (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="text-xs h-7"
                                      onClick={() => openScopeModal(req, false)}
                                    >
                                      Exclude
                                    </Button>
                                  )}
                                </div>
                              )}
                            </TableCell>
                          )}
                        </TableRow>
                      );
                    })}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Scoping Decisions Tab */}
        <TabsContent value="scoping" className="space-y-4 mt-4">
          {scopingDecisions.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <MinusCircle className="h-12 w-12 mb-4 opacity-50" />
                <p className="text-lg font-medium">No scoping decisions</p>
                <p className="text-sm">All requirements are implicitly in-scope</p>
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[120px]">Requirement</TableHead>
                      <TableHead>Title</TableHead>
                      <TableHead>Scope</TableHead>
                      <TableHead>Justification</TableHead>
                      <TableHead>Scoped By</TableHead>
                      {canManage && <TableHead className="w-[80px]">Actions</TableHead>}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {scopingDecisions.map((sd) => (
                      <TableRow key={sd.id}>
                        <TableCell className="font-mono text-sm font-medium">
                          {sd.requirement.identifier}
                        </TableCell>
                        <TableCell className="max-w-xs">
                          <span className="line-clamp-1 text-sm">{sd.requirement.title}</span>
                        </TableCell>
                        <TableCell>
                          {sd.in_scope ? (
                            <Badge className="text-xs bg-green-500/10 text-green-700 dark:text-green-400">In scope</Badge>
                          ) : (
                            <Badge variant="secondary" className="text-xs">Out of scope</Badge>
                          )}
                        </TableCell>
                        <TableCell className="max-w-xs">
                          <span className="line-clamp-2 text-xs text-muted-foreground">
                            {sd.justification || '—'}
                          </span>
                        </TableCell>
                        <TableCell className="text-sm">{sd.scoped_by.name}</TableCell>
                        {canManage && (
                          <TableCell>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-xs h-7"
                              onClick={() => handleResetScope(sd.requirement.id)}
                            >
                              Reset
                            </Button>
                          </TableCell>
                        )}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>

      {/* Scoping Modal */}
      <Dialog open={showScopeModal} onOpenChange={setShowScopeModal}>
        <DialogContent className="sm:max-w-[450px]">
          <DialogHeader>
            <DialogTitle>
              {scopeForm.in_scope ? 'Include' : 'Exclude'} Requirement
            </DialogTitle>
            <DialogDescription>
              {scopeTarget?.identifier}: {scopeTarget?.title}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {scopeError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {scopeError}
              </div>
            )}
            {!scopeForm.in_scope && (
              <div className="space-y-2">
                <Label>Justification (required for exclusion)</Label>
                <Input
                  placeholder="Why is this requirement out of scope?"
                  value={scopeForm.justification}
                  onChange={(e) => setScopeForm((f) => ({ ...f, justification: e.target.value }))}
                />
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowScopeModal(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleSetScope}
              disabled={scopeLoading || (!scopeForm.in_scope && !scopeForm.justification)}
            >
              {scopeLoading ? 'Saving...' : 'Confirm'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
