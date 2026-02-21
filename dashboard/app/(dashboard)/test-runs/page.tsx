'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  Play, ChevronLeft, ChevronRight, Clock, CheckCircle2, XCircle,
  AlertTriangle, Activity, ExternalLink, Loader2,
} from 'lucide-react';
import { TestRun, listTestRuns, createTestRun, cancelTestRun } from '@/lib/api';
import { cn } from '@/lib/utils';

const RUN_STATUS_STYLES: Record<string, { variant: 'default' | 'secondary' | 'destructive' | 'outline'; label: string }> = {
  pending: { variant: 'outline', label: 'Pending' },
  running: { variant: 'default', label: 'Running' },
  completed: { variant: 'secondary', label: 'Completed' },
  failed: { variant: 'destructive', label: 'Failed' },
  cancelled: { variant: 'outline', label: 'Cancelled' },
};

function formatDuration(ms: number | null | undefined): string {
  if (!ms) return '—';
  if (ms < 1000) return `${ms}ms`;
  const secs = ms / 1000;
  if (secs < 60) return `${secs.toFixed(1)}s`;
  const mins = Math.floor(secs / 60);
  const remSecs = Math.round(secs % 60);
  return `${mins}m ${remSecs}s`;
}

export default function TestRunsPage() {
  const { hasRole } = useAuth();
  const canTrigger = hasRole('ciso', 'compliance_manager', 'security_engineer', 'devops_engineer');

  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  const [statusFilter, setStatusFilter] = useState('');
  const [triggerFilter, setTriggerFilter] = useState('');
  const [showTrigger, setShowTrigger] = useState(false);
  const [triggerLoading, setTriggerLoading] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
        sort: 'created_at',
        order: 'desc',
      };
      if (statusFilter) params.status = statusFilter;
      if (triggerFilter) params.trigger_type = triggerFilter;

      const res = await listTestRuns(params);
      setRuns(res.data || []);
      setTotal(res.meta?.total || 0);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, [page, statusFilter, triggerFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  async function handleTriggerRun() {
    setTriggerLoading(true);
    try {
      await createTestRun();
      setShowTrigger(false);
      fetchData();
    } catch {
      // handle
    } finally {
      setTriggerLoading(false);
    }
  }

  async function handleCancel(id: string) {
    if (!confirm('Cancel this test run?')) return;
    try {
      await cancelTestRun(id);
      fetchData();
    } catch {
      // handle
    }
  }

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Activity className="h-8 w-8" />
            Test Execution History
          </h1>
          <p className="text-muted-foreground mt-1">
            View and manage test sweeps across your controls
          </p>
        </div>
        {canTrigger && (
          <Button onClick={() => setShowTrigger(true)}>
            <Play className="h-4 w-4 mr-2" />
            Run All Tests
          </Button>
        )}
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Statuses</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="running">Running</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
            <SelectItem value="failed">Failed</SelectItem>
            <SelectItem value="cancelled">Cancelled</SelectItem>
          </SelectContent>
        </Select>
        <Select value={triggerFilter} onValueChange={(v) => { setTriggerFilter(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="Trigger" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Triggers</SelectItem>
            <SelectItem value="scheduled">Scheduled</SelectItem>
            <SelectItem value="manual">Manual</SelectItem>
            <SelectItem value="on_change">On Change</SelectItem>
            <SelectItem value="webhook">Webhook</SelectItem>
          </SelectContent>
        </Select>
        <span className="text-sm text-muted-foreground">{total} runs</span>
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : runs.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Activity className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">No test runs yet</p>
              <p className="text-sm">Trigger a manual run or wait for scheduled tests</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[80px]">Run #</TableHead>
                  <TableHead className="w-[100px]">Status</TableHead>
                  <TableHead className="w-[100px]">Trigger</TableHead>
                  <TableHead className="w-[80px] text-center">Tests</TableHead>
                  <TableHead className="w-[80px] text-center text-green-600 dark:text-green-400">Pass</TableHead>
                  <TableHead className="w-[80px] text-center text-red-600 dark:text-red-400">Fail</TableHead>
                  <TableHead className="w-[80px] text-center">Errors</TableHead>
                  <TableHead className="w-[90px]">Duration</TableHead>
                  <TableHead className="w-[120px]">Started</TableHead>
                  <TableHead className="w-[100px]">Triggered By</TableHead>
                  <TableHead className="w-[80px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {runs.map((run) => (
                  <TableRow key={run.id}>
                    <TableCell className="font-mono text-sm font-medium">
                      #{run.run_number}
                    </TableCell>
                    <TableCell>
                      <Badge variant={RUN_STATUS_STYLES[run.status]?.variant || 'outline'}>
                        {run.status === 'running' && <Loader2 className="h-3 w-3 mr-1 animate-spin" />}
                        {RUN_STATUS_STYLES[run.status]?.label || run.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {run.trigger_type}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-center font-medium">{run.total_tests}</TableCell>
                    <TableCell className="text-center text-green-600 dark:text-green-400 font-medium">{run.passed}</TableCell>
                    <TableCell className="text-center text-red-600 dark:text-red-400 font-medium">{run.failed}</TableCell>
                    <TableCell className="text-center text-muted-foreground">{run.errors}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDuration(run.duration_ms)}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {run.started_at ? new Date(run.started_at).toLocaleString() : '—'}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {run.triggered_by?.name || (run.trigger_type === 'scheduled' ? 'System' : '—')}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Link href={`/test-runs/${run.id}`}>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <ExternalLink className="h-4 w-4" />
                          </Button>
                        </Link>
                        {['pending', 'running'].includes(run.status) && canTrigger && (
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive"
                            onClick={() => handleCancel(run.id)}>
                            <XCircle className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-muted-foreground">Page {page} of {totalPages}</span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Trigger Run Dialog */}
      <Dialog open={showTrigger} onOpenChange={setShowTrigger}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Run All Active Tests</DialogTitle>
            <DialogDescription>
              This will trigger a full test sweep across all active tests. Only one run is allowed at a time.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowTrigger(false)}>Cancel</Button>
            <Button onClick={handleTriggerRun} disabled={triggerLoading}>
              {triggerLoading ? 'Starting...' : 'Run Tests'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
