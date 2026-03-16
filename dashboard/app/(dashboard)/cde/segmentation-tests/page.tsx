'use client';

import { useState, useEffect, useCallback } from 'react';
import { Plus, Search, MoreVertical, AlertCircle, TestTube2, AlertTriangle } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import {
  SegmentationTest,
  SegmentationTestResult,
  listSegmentationTests,
  createSegmentationTest,
  updateSegmentationTest,
  deleteSegmentationTest,
} from '@/lib/api';
import {
  TEST_RESULT_LABELS,
  TEST_RESULT_COLORS,
} from '@/components/cde/constants';
import { useAuth } from '@/lib/auth-context';

const TEST_RESULTS: SegmentationTestResult[] = ['pass', 'fail', 'pending'];

const emptyForm = {
  name: '',
  description: '',
  result: 'pending' as SegmentationTestResult,
  tester: '',
  tested_at: '',
  next_test_date: '',
  notes: '',
};

/** Returns true if the date is within 30 days from today */
function isAlertDate(dateStr: string | undefined): boolean {
  if (!dateStr) return false;
  const d = new Date(dateStr);
  const now = new Date();
  const diffMs = d.getTime() - now.getTime();
  return diffMs >= 0 && diffMs <= 30 * 24 * 60 * 60 * 1000;
}

/** Returns true if the date is in the past */
function isOverdue(dateStr: string | undefined): boolean {
  if (!dateStr) return false;
  return new Date(dateStr) < new Date();
}

function formatDate(dateStr: string | undefined): string {
  if (!dateStr) return '—';
  return new Date(dateStr).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}

export default function SegmentationTestsPage() {
  const { hasRole } = useAuth();
  const canWrite = hasRole('ciso', 'compliance_manager', 'security_engineer');

  const [tests, setTests] = useState<SegmentationTest[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');
  const [resultFilter, setResultFilter] = useState('');

  const [showDialog, setShowDialog] = useState(false);
  const [editingTest, setEditingTest] = useState<SegmentationTest | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState('');

  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = { page: String(page), per_page: String(perPage) };
      if (search) params.search = search;
      if (resultFilter) params.result = resultFilter;
      const res = await listSegmentationTests(params);
      setTests(res.data);
      setTotal(res.meta?.total ?? res.data.length);
    } catch (err) {
      console.error('Failed to fetch segmentation tests:', err);
    } finally {
      setLoading(false);
    }
  }, [page, search, resultFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  function openCreate() {
    setEditingTest(null);
    setForm(emptyForm);
    setFormError('');
    setShowDialog(true);
  }

  function openEdit(test: SegmentationTest) {
    setEditingTest(test);
    setForm({
      name: test.name,
      description: test.description ?? '',
      result: test.result,
      tester: test.tester ?? '',
      tested_at: test.tested_at ? test.tested_at.slice(0, 10) : '',
      next_test_date: test.next_test_date ? test.next_test_date.slice(0, 10) : '',
      notes: test.notes ?? '',
    });
    setFormError('');
    setShowDialog(true);
  }

  async function handleSubmit() {
    setFormError('');
    setFormLoading(true);
    try {
      if (editingTest) {
        await updateSegmentationTest(editingTest.id, form);
      } else {
        await createSegmentationTest(form);
      }
      setShowDialog(false);
      fetchData();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to save test');
    } finally {
      setFormLoading(false);
    }
  }

  async function handleDelete() {
    if (!deletingId) return;
    setDeleteLoading(true);
    try {
      await deleteSegmentationTest(deletingId);
      setDeletingId(null);
      fetchData();
    } catch (err) {
      console.error('Failed to delete test:', err);
    } finally {
      setDeleteLoading(false);
    }
  }

  const passCount = tests.filter((t) => t.result === 'pass').length;
  const failCount = tests.filter((t) => t.result === 'fail').length;
  const overdueCount = tests.filter((t) => isOverdue(t.next_test_date)).length;
  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Segmentation Tests</h1>
          <p className="text-muted-foreground text-sm mt-1">
            Track penetration and segmentation test results per PCI DSS Req 11.4
          </p>
        </div>
        {canWrite && (
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4 mr-2" />
            Add Test
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold">{total}</div>
            <p className="text-xs text-muted-foreground">Total Tests</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold text-green-600">{passCount}</div>
            <p className="text-xs text-muted-foreground">Passing</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className={`text-2xl font-bold ${failCount > 0 ? 'text-red-600' : 'text-muted-foreground'}`}>
              {failCount}
            </div>
            <p className="text-xs text-muted-foreground">Failing</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className={`text-2xl font-bold ${overdueCount > 0 ? 'text-orange-600' : 'text-muted-foreground'}`}>
              {overdueCount}
            </div>
            <p className="text-xs text-muted-foreground">Overdue for Retest</p>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-4">
            <div className="flex-1 min-w-[200px]">
              <Label className="text-xs">Search</Label>
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search tests..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="w-[160px]">
              <Label className="text-xs">Result</Label>
              <Select value={resultFilter || 'all'} onValueChange={(v) => { setResultFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All results</SelectItem>
                  {TEST_RESULTS.map((r) => (
                    <SelectItem key={r} value={r}>{TEST_RESULT_LABELS[r]}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button variant="outline" size="sm" onClick={() => { setSearch(searchInput); setPage(1); }}>
              Search
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Result</TableHead>
                <TableHead>Tester</TableHead>
                <TableHead>Tested On</TableHead>
                <TableHead>Next Test Date</TableHead>
                {canWrite && <TableHead className="w-[50px]" />}
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={canWrite ? 6 : 5} className="text-center py-8 text-muted-foreground">
                    Loading tests...
                  </TableCell>
                </TableRow>
              ) : tests.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={canWrite ? 6 : 5} className="text-center py-8">
                    <TestTube2 className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                    <p className="text-muted-foreground text-sm">No segmentation tests recorded</p>
                  </TableCell>
                </TableRow>
              ) : (
                tests.map((test) => {
                  const overdue = isOverdue(test.next_test_date);
                  const alertSoon = !overdue && isAlertDate(test.next_test_date);
                  return (
                    <TableRow key={test.id}>
                      <TableCell>
                        <div className="font-medium">{test.name}</div>
                        {test.description && (
                          <div className="text-xs text-muted-foreground line-clamp-1">{test.description}</div>
                        )}
                      </TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${TEST_RESULT_COLORS[test.result]}`}>
                          {TEST_RESULT_LABELS[test.result]}
                        </span>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {test.tester || '—'}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {formatDate(test.tested_at)}
                      </TableCell>
                      <TableCell>
                        {test.next_test_date ? (
                          <div className="flex items-center gap-1.5">
                            <span className={`text-sm ${overdue ? 'text-red-600 font-medium' : alertSoon ? 'text-orange-600 font-medium' : 'text-muted-foreground'}`}>
                              {formatDate(test.next_test_date)}
                            </span>
                            {overdue && (
                              <AlertTriangle className="h-3.5 w-3.5 text-red-500" aria-label="Overdue" />
                            )}
                            {alertSoon && !overdue && (
                              <AlertTriangle className="h-3.5 w-3.5 text-orange-500" aria-label="Due soon" />
                            )}
                          </div>
                        ) : (
                          <span className="text-sm text-muted-foreground">—</span>
                        )}
                      </TableCell>
                      {canWrite && (
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon">
                                <MoreVertical className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => openEdit(test)}>Edit</DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                className="text-destructive"
                                onClick={() => setDeletingId(test.id)}
                              >
                                Delete
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      )}
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} of {total}
          </p>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              Previous
            </Button>
            <span className="text-sm">Page {page} of {totalPages}</span>
            <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              Next
            </Button>
          </div>
        </div>
      )}

      {/* Create / Edit Dialog */}
      <Dialog open={showDialog} onOpenChange={setShowDialog}>
        <DialogContent className="sm:max-w-[520px]">
          <DialogHeader>
            <DialogTitle>{editingTest ? 'Edit Test' : 'Add Segmentation Test'}</DialogTitle>
            <DialogDescription>
              Record a segmentation or penetration test result.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            {formError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
                <AlertCircle className="h-4 w-4 shrink-0" />
                {formError}
              </div>
            )}
            <div className="space-y-2">
              <Label>Test Name <span className="text-destructive">*</span></Label>
              <Input
                placeholder="e.g. Q1 2026 CDE Segmentation Test"
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label>Result <span className="text-destructive">*</span></Label>
              <Select value={form.result} onValueChange={(v) => setForm((f) => ({ ...f, result: v as SegmentationTestResult }))}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {TEST_RESULTS.map((r) => (
                    <SelectItem key={r} value={r}>{TEST_RESULT_LABELS[r]}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Tester</Label>
              <Input
                placeholder="e.g. security@company.com or external firm"
                value={form.tester}
                onChange={(e) => setForm((f) => ({ ...f, tester: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Tested On</Label>
                <Input
                  type="date"
                  value={form.tested_at}
                  onChange={(e) => setForm((f) => ({ ...f, tested_at: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label>Next Test Date</Label>
                <Input
                  type="date"
                  value={form.next_test_date}
                  onChange={(e) => setForm((f) => ({ ...f, next_test_date: e.target.value }))}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Description</Label>
              <Textarea
                placeholder="Scope, methodology, and test objectives..."
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                rows={2}
              />
            </div>
            <div className="space-y-2">
              <Label>Notes</Label>
              <Textarea
                placeholder="Findings, remediation notes, QSA evidence..."
                value={form.notes}
                onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
                rows={2}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDialog(false)}>Cancel</Button>
            <Button onClick={handleSubmit} disabled={formLoading || !form.name}>
              {formLoading ? 'Saving...' : editingTest ? 'Save Changes' : 'Add Test'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deletingId} onOpenChange={() => setDeletingId(null)}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>Delete Test</DialogTitle>
            <DialogDescription>
              This will permanently remove this segmentation test record. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeletingId(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteLoading}>
              {deleteLoading ? 'Deleting...' : 'Delete Test'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
