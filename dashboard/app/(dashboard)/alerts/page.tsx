'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  AlertTriangle, Search, ChevronLeft, ChevronRight, Clock,
  Bell, CheckCircle2, XCircle, Shield, User, ExternalLink,
} from 'lucide-react';
import { Alert, AlertQueueData, getAlertQueue } from '@/lib/api';
import { cn } from '@/lib/utils';

const SEVERITY_COLORS: Record<string, string> = {
  critical: 'bg-red-600 text-white',
  high: 'bg-orange-500 text-white',
  medium: 'bg-amber-500 text-white',
  low: 'bg-blue-500 text-white',
};

const STATUS_STYLES: Record<string, { variant: 'default' | 'secondary' | 'destructive' | 'outline'; label: string }> = {
  open: { variant: 'destructive', label: 'Open' },
  acknowledged: { variant: 'outline', label: 'Acknowledged' },
  in_progress: { variant: 'default', label: 'In Progress' },
  resolved: { variant: 'secondary', label: 'Resolved' },
  suppressed: { variant: 'outline', label: 'Suppressed' },
  closed: { variant: 'secondary', label: 'Closed' },
};

function formatHoursRemaining(hours: number | null | undefined): string {
  if (hours == null) return '—';
  if (hours < 0) return `${Math.abs(hours).toFixed(1)}h overdue`;
  if (hours < 1) return `${Math.round(hours * 60)}m left`;
  return `${hours.toFixed(1)}h left`;
}

export default function AlertQueuePage() {
  const { hasRole } = useAuth();
  const [queueData, setQueueData] = useState<AlertQueueData | null>(null);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [queue, setQueue] = useState('active');
  const [severityFilter, setSeverityFilter] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');

  useEffect(() => {
    const timer = setTimeout(() => { setSearch(searchInput); setPage(1); }, 300);
    return () => clearTimeout(timer);
  }, [searchInput]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
        queue,
      };
      if (severityFilter) params.severity = severityFilter;
      if (search) params.search = search;

      const res = await getAlertQueue(params);
      setQueueData(res.data);
      setTotal(res.meta?.total || 0);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, [page, queue, severityFilter, search]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <Bell className="h-8 w-8" />
          Alert Queue
        </h1>
        <p className="text-muted-foreground mt-1">
          Active alerts sorted by urgency — manage, assign, and resolve
        </p>
      </div>

      {/* Queue summary cards */}
      {queueData?.queue_summary && (
        <div className="grid gap-4 md:grid-cols-5">
          <Card className={cn(queue === 'active' && 'ring-2 ring-primary')}>
            <CardContent className="pt-6 cursor-pointer" onClick={() => { setQueue('active'); setPage(1); }}>
              <div className="text-2xl font-bold text-amber-600 dark:text-amber-400">
                {queueData.queue_summary.active}
              </div>
              <p className="text-xs text-muted-foreground">Active</p>
            </CardContent>
          </Card>
          <Card className={cn(queue === 'resolved' && 'ring-2 ring-primary')}>
            <CardContent className="pt-6 cursor-pointer" onClick={() => { setQueue('resolved'); setPage(1); }}>
              <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                {queueData.queue_summary.resolved}
              </div>
              <p className="text-xs text-muted-foreground">Resolved</p>
            </CardContent>
          </Card>
          <Card className={cn(queue === 'suppressed' && 'ring-2 ring-primary')}>
            <CardContent className="pt-6 cursor-pointer" onClick={() => { setQueue('suppressed'); setPage(1); }}>
              <div className="text-2xl font-bold text-gray-600 dark:text-gray-400">
                {queueData.queue_summary.suppressed}
              </div>
              <p className="text-xs text-muted-foreground">Suppressed</p>
            </CardContent>
          </Card>
          <Card className={cn(queue === 'all' && 'ring-2 ring-primary')}>
            <CardContent className="pt-6 cursor-pointer" onClick={() => { setQueue('all'); setPage(1); }}>
              <div className="text-2xl font-bold">
                {queueData.queue_summary.closed}
              </div>
              <p className="text-xs text-muted-foreground">Closed</p>
            </CardContent>
          </Card>
          <Card className={queueData.queue_summary.sla_breached > 0 ? 'border-red-500/40' : ''}>
            <CardContent className="pt-6">
              <div className={cn(
                'text-2xl font-bold',
                queueData.queue_summary.sla_breached > 0 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400'
              )}>
                {queueData.queue_summary.sla_breached}
              </div>
              <p className="text-xs text-muted-foreground">SLA Breached</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search alerts..."
            className="pl-10"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </div>
        <Select value={severityFilter} onValueChange={(v) => { setSeverityFilter(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="Severity" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Severity</SelectItem>
            <SelectItem value="critical">Critical</SelectItem>
            <SelectItem value="high">High</SelectItem>
            <SelectItem value="medium">Medium</SelectItem>
            <SelectItem value="low">Low</SelectItem>
          </SelectContent>
        </Select>
        <span className="text-sm text-muted-foreground">{total} alerts</span>
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : !queueData?.alerts || queueData.alerts.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <CheckCircle2 className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">No alerts in this queue</p>
              <p className="text-sm">All clear!</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[60px]">#</TableHead>
                  <TableHead className="w-[90px]">Severity</TableHead>
                  <TableHead>Title</TableHead>
                  <TableHead className="w-[100px]">Status</TableHead>
                  <TableHead className="w-[100px]">Control</TableHead>
                  <TableHead className="w-[100px]">Assigned To</TableHead>
                  <TableHead className="w-[100px]">SLA</TableHead>
                  <TableHead className="w-[90px]">Created</TableHead>
                  <TableHead className="w-[50px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {queueData.alerts.map((alert) => (
                  <TableRow key={alert.id} className={alert.sla_breached ? 'bg-red-500/5' : ''}>
                    <TableCell className="font-mono text-sm">
                      #{alert.alert_number}
                    </TableCell>
                    <TableCell>
                      <Badge className={cn('text-xs', SEVERITY_COLORS[alert.severity])}>
                        {alert.severity}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Link href={`/alerts/${alert.id}`} className="hover:text-primary">
                        <span className="text-sm font-medium line-clamp-1">{alert.title}</span>
                      </Link>
                    </TableCell>
                    <TableCell>
                      <Badge variant={STATUS_STYLES[alert.status]?.variant || 'outline'}>
                        {STATUS_STYLES[alert.status]?.label || alert.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {alert.control_identifier || alert.control?.identifier ? (
                        <Badge variant="outline" className="text-[10px] font-mono">
                          {alert.control_identifier || alert.control?.identifier}
                        </Badge>
                      ) : '—'}
                    </TableCell>
                    <TableCell className="text-sm">
                      {alert.assigned_to_name || alert.assigned_to?.name || (
                        <span className="text-muted-foreground text-xs">Unassigned</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {alert.sla_deadline ? (
                        <span className={cn(
                          'text-xs font-medium',
                          alert.sla_breached ? 'text-red-600 dark:text-red-400' :
                          (alert.hours_remaining != null && alert.hours_remaining < 2)
                            ? 'text-amber-600 dark:text-amber-400'
                            : 'text-muted-foreground'
                        )}>
                          {alert.sla_breached && <XCircle className="h-3 w-3 inline mr-0.5" />}
                          {formatHoursRemaining(alert.hours_remaining)}
                        </span>
                      ) : (
                        <span className="text-xs text-muted-foreground">No SLA</span>
                      )}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {new Date(alert.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <Link href={`/alerts/${alert.id}`}>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <ExternalLink className="h-4 w-4" />
                        </Button>
                      </Link>
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
