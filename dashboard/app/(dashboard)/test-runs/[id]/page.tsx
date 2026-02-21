'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  ArrowLeft, Activity, CheckCircle2, XCircle, AlertTriangle,
  Clock, ChevronLeft, ChevronRight, FileText, Loader2,
} from 'lucide-react';
import {
  TestRun, TestResult, getTestRun, listTestRunResults, getTestRunResult,
} from '@/lib/api';
import { cn } from '@/lib/utils';
import Link from 'next/link';

const RESULT_STATUS_STYLES: Record<string, { color: string; icon: typeof CheckCircle2; label: string }> = {
  pass: { color: 'text-green-600 dark:text-green-400', icon: CheckCircle2, label: 'Pass' },
  fail: { color: 'text-red-600 dark:text-red-400', icon: XCircle, label: 'Fail' },
  error: { color: 'text-orange-600 dark:text-orange-400', icon: AlertTriangle, label: 'Error' },
  warning: { color: 'text-amber-600 dark:text-amber-400', icon: AlertTriangle, label: 'Warning' },
  skip: { color: 'text-gray-500', icon: Clock, label: 'Skip' },
};

const SEVERITY_COLORS: Record<string, string> = {
  critical: 'bg-red-600 text-white',
  high: 'bg-orange-500 text-white',
  medium: 'bg-amber-500 text-white',
  low: 'bg-blue-500 text-white',
  informational: 'bg-gray-500 text-white',
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

export default function TestRunDetailPage() {
  const params = useParams();
  const router = useRouter();
  const runId = params.id as string;

  const [run, setRun] = useState<TestRun | null>(null);
  const [results, setResults] = useState<TestResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 50;

  const [selectedResult, setSelectedResult] = useState<TestResult | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [showDetail, setShowDetail] = useState(false);

  // Filters
  const [statusFilter, setStatusFilter] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const resultParams: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
      };
      if (statusFilter) resultParams.status = statusFilter;

      const [runRes, resultsRes] = await Promise.all([
        getTestRun(runId),
        listTestRunResults(runId, resultParams),
      ]);
      setRun(runRes.data);
      setResults(resultsRes.data || []);
      setTotal(resultsRes.meta?.total || 0);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, [runId, page, statusFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  async function openDetail(resultId: string) {
    setDetailLoading(true);
    setShowDetail(true);
    try {
      const res = await getTestRunResult(runId, resultId);
      setSelectedResult(res.data);
    } catch {
      // handle
    } finally {
      setDetailLoading(false);
    }
  }

  const totalPages = Math.ceil(total / perPage);

  if (loading && !run) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <Button variant="ghost" size="sm" className="mb-2" onClick={() => router.push('/test-runs')}>
          <ArrowLeft className="h-4 w-4 mr-1" /> Back to Test Runs
        </Button>
        {run && (
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">Test Run #{run.run_number}</h1>
            <Badge variant={
              run.status === 'completed' ? 'secondary' :
              run.status === 'running' ? 'default' :
              run.status === 'failed' ? 'destructive' : 'outline'
            }>
              {run.status === 'running' && <Loader2 className="h-3 w-3 mr-1 animate-spin" />}
              {run.status}
            </Badge>
            <Badge variant="outline">{run.trigger_type}</Badge>
          </div>
        )}
      </div>

      {/* Run summary cards */}
      {run && (
        <div className="grid gap-4 md:grid-cols-6">
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold">{run.total_tests}</div>
              <p className="text-xs text-muted-foreground">Total Tests</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-green-600 dark:text-green-400">{run.passed}</div>
              <p className="text-xs text-muted-foreground">Passed</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-red-600 dark:text-red-400">{run.failed}</div>
              <p className="text-xs text-muted-foreground">Failed</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">{run.errors}</div>
              <p className="text-xs text-muted-foreground">Errors</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-muted-foreground">{run.skipped + run.warnings}</div>
              <p className="text-xs text-muted-foreground">Skip/Warn</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold">{formatDuration(run.duration_ms)}</div>
              <p className="text-xs text-muted-foreground">Duration</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Results filter */}
      <div className="flex items-center gap-4">
        <span className="text-sm font-medium">Filter results:</span>
        {['', 'fail', 'error', 'pass', 'warning', 'skip'].map((s) => (
          <Button
            key={s || 'all'}
            variant={statusFilter === s ? 'default' : 'outline'}
            size="sm"
            onClick={() => { setStatusFilter(s); setPage(1); }}
          >
            {s || 'All'} {s && RESULT_STATUS_STYLES[s] && (
              <span className="ml-1">
                ({s === 'fail' ? run?.failed : s === 'pass' ? run?.passed :
                  s === 'error' ? run?.errors : s === 'warning' ? run?.warnings : run?.skipped})
              </span>
            )}
          </Button>
        ))}
      </div>

      {/* Results table */}
      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : results.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <FileText className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No results found</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[80px]">Status</TableHead>
                  <TableHead className="w-[80px]">Severity</TableHead>
                  <TableHead className="w-[120px]">Test ID</TableHead>
                  <TableHead>Test</TableHead>
                  <TableHead className="w-[120px]">Control</TableHead>
                  <TableHead>Message</TableHead>
                  <TableHead className="w-[90px]">Duration</TableHead>
                  <TableHead className="w-[80px]">Alert</TableHead>
                  <TableHead className="w-[50px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {results.map((result) => {
                  const style = RESULT_STATUS_STYLES[result.status];
                  const Icon = style?.icon || Clock;
                  return (
                    <TableRow key={result.id} className={result.status === 'fail' ? 'bg-red-500/5' : result.status === 'error' ? 'bg-orange-500/5' : ''}>
                      <TableCell>
                        <div className={cn('flex items-center gap-1', style?.color)}>
                          <Icon className="h-4 w-4" />
                          <span className="text-xs font-medium">{style?.label || result.status}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge className={cn('text-[10px]', SEVERITY_COLORS[result.severity])}>
                          {result.severity}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-mono text-xs">{result.test.identifier}</TableCell>
                      <TableCell className="text-sm">{result.test.title}</TableCell>
                      <TableCell>
                        {result.control ? (
                          <Link href={`/controls/${result.control.id}`}>
                            <Badge variant="outline" className="text-[10px] font-mono cursor-pointer">
                              {result.control.identifier}
                            </Badge>
                          </Link>
                        ) : '—'}
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground max-w-xs truncate">
                        {result.message || '—'}
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {formatDuration(result.duration_ms)}
                      </TableCell>
                      <TableCell>
                        {result.alert_generated ? (
                          result.alert_id ? (
                            <Link href={`/alerts/${result.alert_id}`}>
                              <Badge variant="destructive" className="text-[10px] cursor-pointer">Alert</Badge>
                            </Link>
                          ) : (
                            <Badge variant="destructive" className="text-[10px]">Alert</Badge>
                          )
                        ) : '—'}
                      </TableCell>
                      <TableCell>
                        <Button variant="ghost" size="icon" className="h-8 w-8"
                          onClick={() => openDetail(result.id)}>
                          <FileText className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
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

      {/* Result Detail Dialog */}
      <Dialog open={showDetail} onOpenChange={setShowDetail}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-auto">
          <DialogHeader>
            <DialogTitle>
              {selectedResult ? `${selectedResult.test.identifier} — ${selectedResult.test.title}` : 'Loading...'}
            </DialogTitle>
          </DialogHeader>
          {detailLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : selectedResult ? (
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <Badge className={cn('text-xs', SEVERITY_COLORS[selectedResult.severity])}>
                  {selectedResult.severity}
                </Badge>
                <Badge variant={selectedResult.status === 'pass' ? 'secondary' : 'destructive'}>
                  {selectedResult.status}
                </Badge>
                {selectedResult.duration_ms && (
                  <span className="text-xs text-muted-foreground">
                    <Clock className="h-3 w-3 inline mr-0.5" />{formatDuration(selectedResult.duration_ms)}
                  </span>
                )}
              </div>

              {selectedResult.message && (
                <div>
                  <p className="text-sm font-medium mb-1">Message</p>
                  <p className="text-sm text-muted-foreground">{selectedResult.message}</p>
                </div>
              )}

              {selectedResult.details && Object.keys(selectedResult.details).length > 0 && (
                <div>
                  <p className="text-sm font-medium mb-1">Details</p>
                  <pre className="text-xs bg-muted p-3 rounded-md overflow-auto max-h-48">
                    {JSON.stringify(selectedResult.details, null, 2)}
                  </pre>
                </div>
              )}

              {selectedResult.output_log && (
                <div>
                  <p className="text-sm font-medium mb-1">Output Log</p>
                  <pre className="text-xs bg-muted p-3 rounded-md overflow-auto max-h-64 whitespace-pre-wrap font-mono">
                    {selectedResult.output_log}
                  </pre>
                </div>
              )}

              {selectedResult.error_message && (
                <div>
                  <p className="text-sm font-medium mb-1 text-destructive">Error</p>
                  <p className="text-sm text-destructive">{selectedResult.error_message}</p>
                </div>
              )}

              <div className="text-xs text-muted-foreground space-y-1">
                {selectedResult.started_at && <p>Started: {new Date(selectedResult.started_at).toLocaleString()}</p>}
                {selectedResult.completed_at && <p>Completed: {new Date(selectedResult.completed_at).toLocaleString()}</p>}
              </div>
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
    </div>
  );
}
