'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
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
  FileCheck,
  Search,
  Plus,
  ExternalLink,
  Shield,
  Calendar,
  CheckCircle2,
  AlertCircle,
  XCircle,
} from 'lucide-react';
import {
  Framework,
  OrgFramework,
  FrameworkDetail,
  listFrameworks,
  listOrgFrameworks,
  activateFramework,
  deactivateOrgFramework,
  getFramework,
} from '@/lib/api';

const CATEGORY_LABELS: Record<string, string> = {
  security_privacy: 'Security & Privacy',
  payment: 'Payment',
  data_privacy: 'Data Privacy',
  ai_governance: 'AI Governance',
  industry: 'Industry',
  custom: 'Custom',
};

const CATEGORY_COLORS: Record<string, string> = {
  security_privacy: 'bg-blue-500/10 text-blue-700 dark:text-blue-400',
  payment: 'bg-purple-500/10 text-purple-700 dark:text-purple-400',
  data_privacy: 'bg-green-500/10 text-green-700 dark:text-green-400',
  ai_governance: 'bg-amber-500/10 text-amber-700 dark:text-amber-400',
  industry: 'bg-slate-500/10 text-slate-700 dark:text-slate-400',
  custom: 'bg-rose-500/10 text-rose-700 dark:text-rose-400',
};

function coverageColor(pct: number) {
  if (pct >= 80) return 'text-green-600 dark:text-green-400';
  if (pct >= 50) return 'text-amber-600 dark:text-amber-400';
  return 'text-red-600 dark:text-red-400';
}

export default function FrameworksPage() {
  const { hasRole } = useAuth();
  const canManage = hasRole('ciso', 'compliance_manager');

  const [allFrameworks, setAllFrameworks] = useState<Framework[]>([]);
  const [orgFrameworks, setOrgFrameworks] = useState<OrgFramework[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [activeTab, setActiveTab] = useState('activated');

  // Activation modal
  const [showActivate, setShowActivate] = useState(false);
  const [selectedFramework, setSelectedFramework] = useState<FrameworkDetail | null>(null);
  const [activateForm, setActivateForm] = useState({
    version_id: '',
    target_date: '',
    notes: '',
    seed_controls: true,
  });
  const [activateLoading, setActivateLoading] = useState(false);
  const [activateError, setActivateError] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [fwRes, orgRes] = await Promise.all([
        listFrameworks(),
        listOrgFrameworks(),
      ]);
      setAllFrameworks(fwRes.data || []);
      setOrgFrameworks(orgRes.data || []);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const activatedIds = new Set(orgFrameworks.map((of) => of.framework.id));

  const availableFrameworks = allFrameworks.filter(
    (fw) =>
      !activatedIds.has(fw.id) &&
      (!search || fw.name.toLowerCase().includes(search.toLowerCase()) ||
        fw.identifier.toLowerCase().includes(search.toLowerCase()))
  );

  const filteredOrgFrameworks = orgFrameworks.filter(
    (of) =>
      !search ||
      of.framework.name.toLowerCase().includes(search.toLowerCase()) ||
      of.framework.identifier.toLowerCase().includes(search.toLowerCase())
  );

  async function handleOpenActivate(fw: Framework) {
    setActivateError('');
    setActivateLoading(false);
    try {
      const detail = await getFramework(fw.id);
      setSelectedFramework(detail.data);
      const activeVersion = detail.data.versions?.find((v) => v.status === 'active');
      setActivateForm({
        version_id: activeVersion?.id || detail.data.versions?.[0]?.id || '',
        target_date: '',
        notes: '',
        seed_controls: true,
      });
      setShowActivate(true);
    } catch {
      // handle
    }
  }

  async function handleActivate() {
    if (!selectedFramework) return;
    setActivateError('');
    setActivateLoading(true);
    try {
      await activateFramework({
        framework_id: selectedFramework.id,
        version_id: activateForm.version_id,
        target_date: activateForm.target_date || undefined,
        notes: activateForm.notes || undefined,
        seed_controls: activateForm.seed_controls,
      });
      setShowActivate(false);
      fetchData();
    } catch (err) {
      setActivateError(err instanceof Error ? err.message : 'Failed to activate');
    } finally {
      setActivateLoading(false);
    }
  }

  async function handleDeactivate(orgFwId: string) {
    if (!confirm('Deactivate this framework? Controls and mappings will be preserved.')) return;
    try {
      await deactivateOrgFramework(orgFwId);
      fetchData();
    } catch {
      // handle
    }
  }

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center py-24">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <FileCheck className="h-8 w-8" />
            Compliance Frameworks
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage your organization&apos;s compliance framework activations
          </p>
        </div>
      </div>

      {/* Search */}
      <div className="relative max-w-sm">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search frameworks..."
          className="pl-10"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="activated">
            Activated ({orgFrameworks.length})
          </TabsTrigger>
          <TabsTrigger value="available">
            Available ({availableFrameworks.length})
          </TabsTrigger>
        </TabsList>

        {/* Activated frameworks */}
        <TabsContent value="activated" className="space-y-4 mt-4">
          {filteredOrgFrameworks.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <Shield className="h-12 w-12 mb-4 opacity-50" />
                <p className="text-lg font-medium">No frameworks activated</p>
                <p className="text-sm">Switch to the Available tab to activate a framework</p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {filteredOrgFrameworks.map((of) => (
                <Link
                  key={of.id}
                  href={`/frameworks/${of.id}`}
                  className="block"
                >
                  <Card className="h-full hover:shadow-md transition-shadow cursor-pointer">
                    <CardHeader className="pb-3">
                      <div className="flex items-start justify-between">
                        <div>
                          <CardTitle className="text-lg">{of.framework.name}</CardTitle>
                          <CardDescription>{of.active_version.display_name}</CardDescription>
                        </div>
                        <span
                          className={`text-xs font-medium px-2 py-1 rounded-full ${
                            CATEGORY_COLORS[of.framework.category] || CATEGORY_COLORS.custom
                          }`}
                        >
                          {CATEGORY_LABELS[of.framework.category] || of.framework.category}
                        </span>
                      </div>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      {/* Coverage bar */}
                      <div className="space-y-2">
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-muted-foreground">Coverage</span>
                          <span className={`font-bold ${coverageColor(of.stats.coverage_pct)}`}>
                            {of.stats.coverage_pct.toFixed(1)}%
                          </span>
                        </div>
                        <Progress value={of.stats.coverage_pct} className="h-2" />
                      </div>

                      {/* Stats row */}
                      <div className="grid grid-cols-3 gap-2 text-center text-xs">
                        <div>
                          <div className="font-semibold text-lg">{of.stats.in_scope}</div>
                          <div className="text-muted-foreground">In Scope</div>
                        </div>
                        <div>
                          <div className="font-semibold text-lg text-green-600 dark:text-green-400">
                            {of.stats.mapped}
                          </div>
                          <div className="text-muted-foreground">Mapped</div>
                        </div>
                        <div>
                          <div className="font-semibold text-lg text-red-600 dark:text-red-400">
                            {of.stats.unmapped}
                          </div>
                          <div className="text-muted-foreground">Gaps</div>
                        </div>
                      </div>

                      {/* Target date */}
                      {of.target_date && (
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <Calendar className="h-3 w-3" />
                          Target: {new Date(of.target_date).toLocaleDateString()}
                        </div>
                      )}

                      {/* Status badge */}
                      <div className="flex items-center justify-between">
                        <Badge variant={of.status === 'active' ? 'default' : 'secondary'}>
                          {of.status}
                        </Badge>
                        {canManage && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-xs text-muted-foreground"
                            onClick={(e) => {
                              e.preventDefault();
                              e.stopPropagation();
                              handleDeactivate(of.id);
                            }}
                          >
                            <XCircle className="h-3 w-3 mr-1" />
                            Deactivate
                          </Button>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          )}
        </TabsContent>

        {/* Available frameworks */}
        <TabsContent value="available" className="space-y-4 mt-4">
          {availableFrameworks.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <CheckCircle2 className="h-12 w-12 mb-4 opacity-50" />
                <p className="text-lg font-medium">All frameworks activated</p>
                <p className="text-sm">Every available framework is already active</p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {availableFrameworks.map((fw) => (
                <Card key={fw.id} className="h-full">
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div>
                        <CardTitle className="text-lg">{fw.name}</CardTitle>
                        <CardDescription className="line-clamp-2">{fw.description}</CardDescription>
                      </div>
                      <span
                        className={`text-xs font-medium px-2 py-1 rounded-full whitespace-nowrap ${
                          CATEGORY_COLORS[fw.category] || CATEGORY_COLORS.custom
                        }`}
                      >
                        {CATEGORY_LABELS[fw.category] || fw.category}
                      </span>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                      <span>{fw.versions_count} version{fw.versions_count !== 1 ? 's' : ''}</span>
                      {fw.website_url && (
                        <a
                          href={fw.website_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1 hover:text-primary"
                          onClick={(e) => e.stopPropagation()}
                        >
                          <ExternalLink className="h-3 w-3" />
                          Website
                        </a>
                      )}
                    </div>
                    {canManage && (
                      <Button
                        className="w-full"
                        onClick={() => handleOpenActivate(fw)}
                      >
                        <Plus className="h-4 w-4 mr-2" />
                        Activate Framework
                      </Button>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>

      {/* Activation Modal */}
      <Dialog open={showActivate} onOpenChange={setShowActivate}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Activate {selectedFramework?.name}</DialogTitle>
            <DialogDescription>
              {selectedFramework?.description}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {activateError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
                <AlertCircle className="h-4 w-4" />
                {activateError}
              </div>
            )}

            <div className="space-y-2">
              <Label>Version</Label>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={activateForm.version_id}
                onChange={(e) => setActivateForm((f) => ({ ...f, version_id: e.target.value }))}
              >
                {selectedFramework?.versions?.map((v) => (
                  <option key={v.id} value={v.id}>
                    {v.display_name} ({v.total_requirements} requirements)
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-2">
              <Label>Target Compliance Date (optional)</Label>
              <Input
                type="date"
                value={activateForm.target_date}
                onChange={(e) => setActivateForm((f) => ({ ...f, target_date: e.target.value }))}
              />
            </div>

            <div className="space-y-2">
              <Label>Notes (optional)</Label>
              <Input
                placeholder="Why are you activating this framework?"
                value={activateForm.notes}
                onChange={(e) => setActivateForm((f) => ({ ...f, notes: e.target.value }))}
              />
            </div>

            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="seed-controls"
                checked={activateForm.seed_controls}
                onChange={(e) => setActivateForm((f) => ({ ...f, seed_controls: e.target.checked }))}
                className="h-4 w-4 rounded border-input"
              />
              <Label htmlFor="seed-controls" className="text-sm font-normal">
                Seed pre-built controls from template library
              </Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowActivate(false)}>
              Cancel
            </Button>
            <Button onClick={handleActivate} disabled={activateLoading || !activateForm.version_id}>
              {activateLoading ? 'Activating...' : 'Activate'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
