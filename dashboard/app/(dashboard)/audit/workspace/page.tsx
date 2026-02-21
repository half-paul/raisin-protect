'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  ClipboardCheck, AlertTriangle, Clock, Eye, FileCheck, Search as SearchIcon,
  BarChart3,
} from 'lucide-react';
import {
  Audit, AuditDashboard,
  listAudits, getAuditDashboard,
} from '@/lib/api';
import {
  AUDIT_STATUS_LABELS, AUDIT_STATUS_COLORS, AUDIT_TYPE_LABELS,
  FINDING_SEVERITY_COLORS,
} from '@/components/audit/constants';

export default function AuditorWorkspacePage() {
  const { user, hasRole } = useAuth();
  const isAuditor = hasRole('auditor');

  const [audits, setAudits] = useState<Audit[]>([]);
  const [dashboard, setDashboard] = useState<AuditDashboard | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      // For auditors, the backend automatically filters to assigned audits
      const [auditsRes, dashRes] = await Promise.all([
        listAudits({ per_page: '100' }),
        getAuditDashboard().catch(() => ({ data: null })),
      ]);
      setAudits(auditsRes.data);
      if (dashRes.data) setDashboard(dashRes.data);
    } catch (err) {
      console.error('Failed to fetch workspace:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const activeAudits = audits.filter(a => !['completed', 'cancelled'].includes(a.status));
  const completedAudits = audits.filter(a => a.status === 'completed');
  const ds = dashboard?.summary;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold">Auditor Workspace</h1>
        <p className="text-sm text-muted-foreground">
          {isAuditor ? 'Your assigned audit engagements' : 'Audit engagements overview'}
        </p>
      </div>

      {/* Summary */}
      {ds && (
        <div className="grid gap-4 md:grid-cols-4">
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{activeAudits.length}</div>
              <p className="text-xs text-muted-foreground">Active Engagements</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-blue-600">{ds.total_open_requests}</div>
              <p className="text-xs text-muted-foreground">Open Requests</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-red-600">{ds.total_open_findings}</div>
              <p className="text-xs text-muted-foreground">Open Findings</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-orange-600">{ds.total_overdue_requests}</div>
              <p className="text-xs text-muted-foreground">Overdue Requests</p>
            </CardContent>
          </Card>
        </div>
      )}

      {loading ? (
        <div className="text-center py-16 text-muted-foreground">Loading...</div>
      ) : (
        <>
          {/* Active Engagements */}
          {activeAudits.length > 0 && (
            <div>
              <h2 className="text-lg font-semibold mb-4">Active Engagements</h2>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {activeAudits.map((audit) => (
                  <Link key={audit.id} href={`/audit/${audit.id}`}>
                    <Card className="hover:border-primary/50 transition-colors cursor-pointer h-full">
                      <CardHeader className="pb-2">
                        <div className="flex items-center justify-between">
                          <CardTitle className="text-sm font-medium truncate">{audit.title}</CardTitle>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium shrink-0 ${AUDIT_STATUS_COLORS[audit.status]}`}>
                            {AUDIT_STATUS_LABELS[audit.status]}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <Badge variant="outline" className="text-[10px]">{AUDIT_TYPE_LABELS[audit.audit_type] || audit.audit_type}</Badge>
                          {audit.audit_firm && <span>{audit.audit_firm}</span>}
                        </div>
                      </CardHeader>
                      <CardContent className="pb-4">
                        {/* Stats */}
                        <div className="grid grid-cols-2 gap-3 text-xs mb-3">
                          <div className="flex items-center gap-1.5">
                            <ClipboardCheck className="h-3 w-3 text-muted-foreground" />
                            <span>{audit.open_requests} open / {audit.total_requests} requests</span>
                          </div>
                          <div className="flex items-center gap-1.5">
                            <AlertTriangle className="h-3 w-3 text-muted-foreground" />
                            <span>{audit.open_findings} open / {audit.total_findings} findings</span>
                          </div>
                        </div>
                        {/* Timeline */}
                        {audit.planned_end && (
                          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <Clock className="h-3 w-3" />
                            <span>Due: {new Date(audit.planned_end).toLocaleDateString()}</span>
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            </div>
          )}

          {/* Overdue items requiring attention */}
          {dashboard?.overdue_requests && dashboard.overdue_requests.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-orange-500" />
                  Overdue Evidence Requests
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Request</TableHead>
                      <TableHead>Audit</TableHead>
                      <TableHead>Assigned To</TableHead>
                      <TableHead>Priority</TableHead>
                      <TableHead className="text-right">Overdue</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {dashboard.overdue_requests.map((req) => (
                      <TableRow key={req.id}>
                        <TableCell className="font-medium text-sm">{req.title}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{req.audit_title}</TableCell>
                        <TableCell className="text-sm">{req.assigned_to_name || '—'}</TableCell>
                        <TableCell>
                          <Badge variant="outline" className="text-xs">{req.priority}</Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <Badge variant="destructive" className="text-xs">{req.days_overdue}d</Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {/* Critical findings */}
          {dashboard?.critical_findings && dashboard.critical_findings.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-red-500" />
                  Critical / High Findings
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Finding</TableHead>
                      <TableHead>Audit</TableHead>
                      <TableHead>Severity</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Owner</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {dashboard.critical_findings.map((f) => (
                      <TableRow key={f.id}>
                        <TableCell className="font-medium text-sm">{f.title}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{f.audit_title}</TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${FINDING_SEVERITY_COLORS[f.severity]}`}>
                            {f.severity}
                          </span>
                        </TableCell>
                        <TableCell className="text-sm">{f.status.replace(/_/g, ' ')}</TableCell>
                        <TableCell className="text-sm">{f.remediation_owner_name || '—'}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {/* Recent activity */}
          {dashboard?.recent_activity && dashboard.recent_activity.length > 0 && (
            <Card>
              <CardHeader><CardTitle className="text-base">Recent Activity</CardTitle></CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {dashboard.recent_activity.slice(0, 10).map((a, i) => (
                    <div key={i} className="flex items-center gap-3 text-sm">
                      <div className="w-2 h-2 rounded-full bg-primary shrink-0" />
                      <div className="flex-1">
                        <span className="font-medium">{a.actor_name}</span>
                        {' '}
                        <span className="text-muted-foreground">{a.type.replace(/_/g, ' ')}</span>
                        {' '}
                        <span>{a.title}</span>
                        <span className="text-muted-foreground"> in {a.audit_title}</span>
                      </div>
                      <span className="text-xs text-muted-foreground shrink-0">{new Date(a.timestamp).toLocaleString()}</span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Completed Engagements */}
          {completedAudits.length > 0 && (
            <Card>
              <CardHeader><CardTitle className="text-base">Completed Engagements ({completedAudits.length})</CardTitle></CardHeader>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Title</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Firm</TableHead>
                      <TableHead>Completed</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {completedAudits.map((a) => (
                      <TableRow key={a.id}>
                        <TableCell className="font-medium text-sm">{a.title}</TableCell>
                        <TableCell><Badge variant="outline" className="text-xs">{AUDIT_TYPE_LABELS[a.audit_type]}</Badge></TableCell>
                        <TableCell className="text-sm">{a.audit_firm || '—'}</TableCell>
                        <TableCell className="text-sm">{a.actual_end ? new Date(a.actual_end).toLocaleDateString() : '—'}</TableCell>
                        <TableCell className="text-right">
                          <Link href={`/audit/${a.id}`}>
                            <Button variant="ghost" size="icon" className="h-8 w-8"><Eye className="h-4 w-4" /></Button>
                          </Link>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {audits.length === 0 && (
            <div className="text-center py-16 text-muted-foreground">
              {isAuditor ? 'No audits assigned to you yet.' : 'No audits found.'}
            </div>
          )}
        </>
      )}
    </div>
  );
}
