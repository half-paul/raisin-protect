'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  AlertTriangle, XCircle, Clock, ArrowLeft, Shield,
  ChevronLeft, ChevronRight, FileText,
} from 'lucide-react';
import {
  StalenessAlert, StalenessSummary,
  getStalenessAlerts,
} from '@/lib/api';
import {
  EVIDENCE_TYPE_LABELS, formatFileSize,
} from '@/components/evidence/freshness-badge';

export default function StalenessPage() {
  const router = useRouter();
  const [summary, setSummary] = useState<StalenessSummary | null>(null);
  const [alerts, setAlerts] = useState<StalenessAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [alertLevel, setAlertLevel] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [daysAhead, setDaysAhead] = useState('30');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
        sort: 'expires_at',
        order: 'asc',
      };
      if (alertLevel) params.alert_level = alertLevel;
      if (typeFilter) params.evidence_type = typeFilter;
      if (daysAhead) params.days_ahead = daysAhead;

      const res = await getStalenessAlerts(params);
      setSummary(res.data.summary);
      setAlerts(res.data.alerts || []);
      setTotal(res.meta?.total || res.data.summary?.total_alerts || 0);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, [page, alertLevel, typeFilter, daysAhead]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <Button variant="ghost" size="sm" className="mb-2" onClick={() => router.push('/evidence')}>
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Evidence Library
        </Button>
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <AlertTriangle className="h-8 w-8 text-amber-600 dark:text-amber-400" />
          Staleness Alerts
        </h1>
        <p className="text-muted-foreground mt-1">
          Evidence artifacts that are expired or expiring soon
        </p>
      </div>

      {/* Summary cards */}
      {summary && (
        <div className="grid gap-4 md:grid-cols-4">
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold">{summary.total_alerts}</div>
              <p className="text-xs text-muted-foreground">Total Alerts</p>
            </CardContent>
          </Card>
          <Card className="border-red-500/20">
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-red-600 dark:text-red-400 flex items-center gap-2">
                <XCircle className="h-5 w-5" />
                {summary.expired}
              </div>
              <p className="text-xs text-muted-foreground">Expired</p>
            </CardContent>
          </Card>
          <Card className="border-amber-500/20">
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-amber-600 dark:text-amber-400 flex items-center gap-2">
                <Clock className="h-5 w-5" />
                {summary.expiring_soon}
              </div>
              <p className="text-xs text-muted-foreground">Expiring Soon</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-blue-600 dark:text-blue-400 flex items-center gap-2">
                <Shield className="h-5 w-5" />
                {summary.affected_controls}
              </div>
              <p className="text-xs text-muted-foreground">Affected Controls</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <Select value={alertLevel} onValueChange={(v) => { setAlertLevel(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Alert Level" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Alerts</SelectItem>
            <SelectItem value="expired">Expired</SelectItem>
            <SelectItem value="expiring_soon">Expiring Soon</SelectItem>
          </SelectContent>
        </Select>
        <Select value={typeFilter} onValueChange={(v) => { setTypeFilter(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Evidence Type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            {Object.entries(EVIDENCE_TYPE_LABELS).map(([k, v]) => (
              <SelectItem key={k} value={k}>{v}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Look ahead:</span>
          <Input
            type="number"
            className="w-[80px]"
            value={daysAhead}
            onChange={(e) => setDaysAhead(e.target.value)}
          />
          <span className="text-sm text-muted-foreground">days</span>
        </div>
        <span className="text-sm text-muted-foreground">{total} alerts</span>
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : alerts.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <AlertTriangle className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">No staleness alerts</p>
              <p className="text-sm">All evidence is fresh. Great job!</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[90px]">Alert</TableHead>
                  <TableHead>Title</TableHead>
                  <TableHead className="w-[120px]">Type</TableHead>
                  <TableHead className="w-[100px]">Collected</TableHead>
                  <TableHead className="w-[100px]">Expires</TableHead>
                  <TableHead className="w-[100px]">Urgency</TableHead>
                  <TableHead>Affected Controls</TableHead>
                  <TableHead className="w-[100px]">Uploaded By</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {alerts.map(alert => (
                  <TableRow key={alert.id} className={alert.alert_level === 'expired' ? 'bg-red-500/5' : 'bg-amber-500/5'}>
                    <TableCell>
                      {alert.alert_level === 'expired' ? (
                        <Badge variant="destructive" className="text-xs">
                          <XCircle className="h-3 w-3 mr-1" />Expired
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="text-xs bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20">
                          <Clock className="h-3 w-3 mr-1" />Expiring
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <Link href={`/evidence/${alert.id}`} className="hover:text-primary">
                        <div className="flex items-center gap-2">
                          <FileText className="h-4 w-4 text-muted-foreground shrink-0" />
                          <span className="text-sm font-medium line-clamp-1">{alert.title}</span>
                        </div>
                      </Link>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {EVIDENCE_TYPE_LABELS[alert.evidence_type] || alert.evidence_type}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-xs">
                      {new Date(alert.collection_date).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-xs">
                      {new Date(alert.expires_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      {alert.alert_level === 'expired' ? (
                        <span className="text-sm font-medium text-red-600 dark:text-red-400">
                          {alert.days_overdue}d overdue
                        </span>
                      ) : (
                        <span className="text-sm font-medium text-amber-600 dark:text-amber-400">
                          {alert.days_until_expiry}d left
                        </span>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {alert.linked_controls.slice(0, 3).map(ctrl => (
                          <Link key={ctrl.id} href={`/controls/${ctrl.id}`}>
                            <Badge variant="secondary" className="text-[10px] cursor-pointer hover:bg-secondary/80">
                              {ctrl.identifier}
                            </Badge>
                          </Link>
                        ))}
                        {alert.linked_controls_count > 3 && (
                          <Badge variant="secondary" className="text-[10px]">
                            +{alert.linked_controls_count - 3}
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {alert.uploaded_by?.name || '—'}
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
    </div>
  );
}
