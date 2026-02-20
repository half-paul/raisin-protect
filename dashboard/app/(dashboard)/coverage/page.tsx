'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  BarChart3,
  Shield,
  CheckCircle2,
  AlertCircle,
  Target,
  TrendingUp,
  ArrowRight,
} from 'lucide-react';
import {
  ControlStats,
  OrgFramework,
  getControlStats,
  listOrgFrameworks,
} from '@/lib/api';

function coverageColor(pct: number) {
  if (pct >= 80) return 'text-green-600 dark:text-green-400';
  if (pct >= 50) return 'text-amber-600 dark:text-amber-400';
  return 'text-red-600 dark:text-red-400';
}

function coverageBg(pct: number) {
  if (pct >= 80) return 'bg-green-500';
  if (pct >= 50) return 'bg-amber-500';
  return 'bg-red-500';
}

export default function CoverageDashboard() {
  const [stats, setStats] = useState<ControlStats | null>(null);
  const [orgFrameworks, setOrgFrameworks] = useState<OrgFramework[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [statsRes, ofRes] = await Promise.all([
        getControlStats(),
        listOrgFrameworks({ status: 'active' }),
      ]);
      setStats(statsRes.data);
      setOrgFrameworks(ofRes.data || []);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center py-24">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  // Compute overall coverage
  const totalInScope = stats?.frameworks_coverage?.reduce((s, f) => s + f.in_scope, 0) || 0;
  const totalCovered = stats?.frameworks_coverage?.reduce((s, f) => s + f.covered, 0) || 0;
  const overallPct = totalInScope > 0 ? (totalCovered / totalInScope) * 100 : 0;
  const totalGaps = stats?.frameworks_coverage?.reduce((s, f) => s + f.gaps, 0) || 0;

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <BarChart3 className="h-8 w-8" />
          Compliance Coverage
        </h1>
        <p className="text-muted-foreground mt-1">
          Overall compliance posture across all activated frameworks
        </p>
      </div>

      {/* Top-level stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Overall Coverage</p>
                <div className={`text-3xl font-bold ${coverageColor(overallPct)}`}>
                  {overallPct.toFixed(1)}%
                </div>
              </div>
              <Target className="h-10 w-10 text-muted-foreground/30" />
            </div>
            <Progress value={overallPct} className="h-2 mt-3" />
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Active Frameworks</p>
                <div className="text-3xl font-bold">
                  {orgFrameworks.length}
                </div>
              </div>
              <Shield className="h-10 w-10 text-muted-foreground/30" />
            </div>
            <p className="text-xs text-muted-foreground mt-3">
              {totalInScope} total in-scope requirements
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Requirements Covered</p>
                <div className="text-3xl font-bold text-green-600 dark:text-green-400">
                  {totalCovered}
                </div>
              </div>
              <CheckCircle2 className="h-10 w-10 text-green-500/30" />
            </div>
            <p className="text-xs text-muted-foreground mt-3">
              of {totalInScope} in scope
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Coverage Gaps</p>
                <div className="text-3xl font-bold text-red-600 dark:text-red-400">
                  {totalGaps}
                </div>
              </div>
              <AlertCircle className="h-10 w-10 text-red-500/30" />
            </div>
            <p className="text-xs text-muted-foreground mt-3">
              requirements needing controls
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Per-framework coverage */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5" />
            Framework Coverage Breakdown
          </CardTitle>
          <CardDescription>
            Coverage posture for each activated framework
          </CardDescription>
        </CardHeader>
        <CardContent>
          {orgFrameworks.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No active frameworks. Activate frameworks to see coverage data.
            </div>
          ) : (
            <div className="space-y-6">
              {orgFrameworks.map((of) => {
                const pct = of.stats.coverage_pct;
                return (
                  <div key={of.id} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Link
                          href={`/frameworks/${of.id}`}
                          className="font-medium hover:text-primary flex items-center gap-1"
                        >
                          {of.framework.name}
                          <ArrowRight className="h-3 w-3 opacity-0 group-hover:opacity-100" />
                        </Link>
                        <Badge variant="outline" className="text-xs">
                          {of.active_version.display_name}
                        </Badge>
                      </div>
                      <div className="flex items-center gap-4 text-sm">
                        <span className="text-muted-foreground">
                          {of.stats.mapped}/{of.stats.in_scope} reqs
                        </span>
                        <span className={`font-bold ${coverageColor(pct)}`}>
                          {pct.toFixed(1)}%
                        </span>
                      </div>
                    </div>
                    <div className="h-3 rounded-full bg-secondary overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all ${coverageBg(pct)}`}
                        style={{ width: `${Math.min(pct, 100)}%` }}
                      />
                    </div>
                    <div className="flex gap-4 text-xs text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <CheckCircle2 className="h-3 w-3 text-green-500" />
                        {of.stats.mapped} covered
                      </span>
                      <span className="flex items-center gap-1">
                        <AlertCircle className="h-3 w-3 text-red-500" />
                        {of.stats.unmapped} gaps
                      </span>
                      <span>
                        {of.stats.out_of_scope} out of scope
                      </span>
                      {of.target_date && (
                        <span>Target: {new Date(of.target_date).toLocaleDateString()}</span>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Control distribution */}
      {stats && (
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Controls by Status</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {Object.entries(stats.by_status || {}).map(([status, count]) => (
                  <div key={status} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Badge
                        variant={
                          status === 'active' ? 'default' :
                          status === 'deprecated' ? 'destructive' : 'secondary'
                        }
                        className="text-xs"
                      >
                        {status}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="w-24 h-2 rounded-full bg-secondary overflow-hidden">
                        <div
                          className="h-full rounded-full bg-primary"
                          style={{ width: `${(count / stats.total) * 100}%` }}
                        />
                      </div>
                      <span className="text-sm font-medium w-8 text-right">{count}</span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Controls by Category</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {Object.entries(stats.by_category || {}).map(([cat, count]) => (
                  <div key={cat} className="flex items-center justify-between">
                    <span className="text-sm capitalize">{cat}</span>
                    <div className="flex items-center gap-2">
                      <div className="w-24 h-2 rounded-full bg-secondary overflow-hidden">
                        <div
                          className="h-full rounded-full bg-primary"
                          style={{ width: `${(count / stats.total) * 100}%` }}
                        />
                      </div>
                      <span className="text-sm font-medium w-8 text-right">{count}</span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
