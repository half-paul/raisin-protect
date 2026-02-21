'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  AlertTriangle, Shield, BarChart3, CheckCircle2, XCircle,
} from 'lucide-react';
import {
  PolicyGapSummary, PolicyGapItem, PolicyGapByFramework, OrgFramework,
  getPolicyGap, getPolicyGapByFramework, listOrgFrameworks,
} from '@/lib/api';

const CATEGORY_LABELS: Record<string, string> = {
  information_security: 'Information Security', access_control: 'Access Control',
  incident_response: 'Incident Response', data_privacy: 'Data Privacy',
  network_security: 'Network Security', encryption: 'Encryption',
  vulnerability_management: 'Vulnerability Management', change_management: 'Change Management',
  business_continuity: 'Business Continuity', secure_development: 'Secure Development',
  vendor_management: 'Vendor Management', acceptable_use: 'Acceptable Use',
  physical_security: 'Physical Security', hr_security: 'HR Security',
  asset_management: 'Asset Management',
  technical: 'Technical', administrative: 'Administrative', operational: 'Operational',
};

export default function PolicyGapDashboardPage() {
  const { hasRole } = useAuth();
  const canView = hasRole('ciso', 'compliance_manager', 'security_engineer', 'auditor');

  const [summary, setSummary] = useState<PolicyGapSummary | null>(null);
  const [gaps, setGaps] = useState<PolicyGapItem[]>([]);
  const [frameworkGaps, setFrameworkGaps] = useState<PolicyGapByFramework[]>([]);
  const [frameworks, setFrameworks] = useState<OrgFramework[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('controls');

  // Filters
  const [frameworkFilter, setFrameworkFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [includePartial, setIncludePartial] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {};
      if (frameworkFilter) params.framework_id = frameworkFilter;
      if (categoryFilter) params.category = categoryFilter;
      if (includePartial) params.include_partial = 'true';

      const [gapRes, fwGapRes, fwRes] = await Promise.all([
        getPolicyGap(params),
        getPolicyGapByFramework(),
        listOrgFrameworks(),
      ]);
      setSummary(gapRes.data.summary);
      setGaps(gapRes.data.gaps);
      setFrameworkGaps(fwGapRes.data);
      setFrameworks(fwRes.data);
    } catch (err) {
      console.error('Failed to fetch gap data:', err);
    } finally {
      setLoading(false);
    }
  }, [frameworkFilter, categoryFilter, includePartial]);

  useEffect(() => { fetchData(); }, [fetchData]);

  if (!canView) {
    return <div className="flex items-center justify-center py-20 text-muted-foreground">Access denied</div>;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <AlertTriangle className="h-6 w-6" /> Policy Gap Analysis
        </h1>
        <p className="text-sm text-muted-foreground">Identify controls without adequate policy coverage</p>
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid gap-4 md:grid-cols-5">
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{summary.total_active_controls}</div>
              <p className="text-xs text-muted-foreground">Total Active Controls</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-green-600">{summary.controls_with_full_coverage}</div>
              <p className="text-xs text-muted-foreground">Full Coverage</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-yellow-600">{summary.controls_with_partial_coverage}</div>
              <p className="text-xs text-muted-foreground">Partial Coverage</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-red-600">{summary.controls_without_coverage}</div>
              <p className="text-xs text-muted-foreground">No Coverage</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{summary.coverage_percentage.toFixed(1)}%</div>
              <Progress value={summary.coverage_percentage} className="mt-2" />
              <p className="text-xs text-muted-foreground mt-1">Policy Coverage</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="controls">
            <Shield className="h-4 w-4 mr-1" /> By Control
          </TabsTrigger>
          <TabsTrigger value="frameworks">
            <BarChart3 className="h-4 w-4 mr-1" /> By Framework
          </TabsTrigger>
        </TabsList>

        {/* By Control Tab */}
        <TabsContent value="controls">
          {/* Filters */}
          <Card className="mb-4">
            <CardContent className="p-4">
              <div className="flex flex-wrap items-end gap-4">
                <div className="w-[180px]">
                  <label className="text-xs font-medium mb-1 block">Framework</label>
                  <Select value={frameworkFilter} onValueChange={(v) => setFrameworkFilter(v === 'all' ? '' : v)}>
                    <SelectTrigger><SelectValue placeholder="All frameworks" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All frameworks</SelectItem>
                      {frameworks.map(fw => (
                        <SelectItem key={fw.id} value={fw.framework.id}>{fw.framework.name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="w-[180px]">
                  <label className="text-xs font-medium mb-1 block">Category</label>
                  <Select value={categoryFilter} onValueChange={(v) => setCategoryFilter(v === 'all' ? '' : v)}>
                    <SelectTrigger><SelectValue placeholder="All categories" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All categories</SelectItem>
                      {Object.entries(CATEGORY_LABELS).map(([k, v]) => (
                        <SelectItem key={k} value={k}>{v}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <Button
                  variant={includePartial ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setIncludePartial(!includePartial)}
                >
                  {includePartial ? 'Including Partial' : 'Show Partial'}
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Controls Without Policy Coverage</CardTitle>
              <CardDescription>{gaps.length} gaps found — ordered by impact (most framework requirements first)</CardDescription>
            </CardHeader>
            <CardContent className="p-0">
              {loading ? (
                <div className="p-6 text-center text-muted-foreground">Loading...</div>
              ) : gaps.length === 0 ? (
                <div className="p-6 text-center">
                  <CheckCircle2 className="h-8 w-8 text-green-500 mx-auto mb-2" />
                  <p className="text-muted-foreground">All controls have policy coverage!</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Control</TableHead>
                      <TableHead>Category</TableHead>
                      <TableHead>Coverage</TableHead>
                      <TableHead>Frameworks</TableHead>
                      <TableHead>Requirements</TableHead>
                      <TableHead>Suggested Policy</TableHead>
                      <TableHead>Owner</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {gaps.map((gap) => (
                      <TableRow key={gap.control.id}>
                        <TableCell>
                          <Link href={`/controls/${gap.control.id}`} className="text-primary hover:underline">
                            <span className="font-mono text-xs mr-2">{gap.control.identifier}</span>
                            <span className="text-sm">{gap.control.title}</span>
                          </Link>
                        </TableCell>
                        <TableCell><Badge variant="outline" className="text-xs">{CATEGORY_LABELS[gap.control.category] || gap.control.category}</Badge></TableCell>
                        <TableCell>
                          <Badge variant={gap.policy_coverage === 'none' ? 'destructive' : 'secondary'} className="text-xs">
                            {gap.policy_coverage === 'none' ? (
                              <><XCircle className="h-3 w-3 mr-1" /> None</>
                            ) : 'Partial'}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-xs">{gap.mapped_frameworks.join(', ')}</TableCell>
                        <TableCell className="text-sm font-medium">{gap.mapped_requirements_count}</TableCell>
                        <TableCell>
                          <div className="flex flex-wrap gap-1">
                            {gap.suggested_categories.map(cat => (
                              <Badge key={cat} variant="secondary" className="text-xs">
                                {CATEGORY_LABELS[cat] || cat}
                              </Badge>
                            ))}
                          </div>
                        </TableCell>
                        <TableCell className="text-sm">{gap.control.owner?.name || '—'}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* By Framework Tab */}
        <TabsContent value="frameworks">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {loading ? (
              <div className="col-span-full text-center py-12 text-muted-foreground">Loading...</div>
            ) : frameworkGaps.length === 0 ? (
              <div className="col-span-full text-center py-12 text-muted-foreground">No framework data available</div>
            ) : (
              frameworkGaps.map((fw) => (
                <Card key={fw.framework.id}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-base">{fw.framework.name}</CardTitle>
                    <CardDescription className="text-xs">{fw.framework.version}</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-2xl font-bold">{fw.policy_coverage_percentage.toFixed(1)}%</span>
                      <Badge variant={fw.gap_count > 0 ? 'destructive' : 'default'} className="text-xs">
                        {fw.gap_count} gaps
                      </Badge>
                    </div>
                    <Progress value={fw.policy_coverage_percentage} className="h-2" />
                    <div className="grid grid-cols-2 gap-2 text-xs">
                      <div>
                        <span className="text-muted-foreground">Total Requirements:</span>
                        <span className="ml-1 font-medium">{fw.total_requirements}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">With Controls:</span>
                        <span className="ml-1 font-medium">{fw.requirements_with_controls}</span>
                      </div>
                      <div>
                        <span className="text-green-600">Covered:</span>
                        <span className="ml-1 font-medium">{fw.controls_with_policy_coverage}</span>
                      </div>
                      <div>
                        <span className="text-red-600">Uncovered:</span>
                        <span className="ml-1 font-medium">{fw.controls_without_policy_coverage}</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
