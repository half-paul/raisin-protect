'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  ArrowLeft, BarChart3, AlertTriangle, CheckCircle, Clock, Target,
} from 'lucide-react';
import {
  Audit, AuditReadiness, AuditStats,
  listAudits, getAuditReadiness, getAuditStats,
} from '@/lib/api';
import {
  AUDIT_STATUS_LABELS, AUDIT_STATUS_COLORS,
  AUDIT_TYPE_LABELS,
} from '@/components/audit/constants';

export default function AuditReadinessPage() {
  const [audits, setAudits] = useState<Audit[]>([]);
  const [selectedAuditId, setSelectedAuditId] = useState('');
  const [readiness, setReadiness] = useState<AuditReadiness | null>(null);
  const [stats, setStats] = useState<AuditStats | null>(null);
  const [loading, setLoading] = useState(true);

  // Fetch active audits
  useEffect(() => {
    (async () => {
      try {
        setLoading(true);
        const res = await listAudits({ per_page: '100' });
        const active = res.data.filter(a => !['completed', 'cancelled'].includes(a.status));
        setAudits(active);
        if (active.length > 0) setSelectedAuditId(active[0].id);
      } catch (err) {
        console.error('Failed to fetch audits:', err);
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  // Fetch readiness for selected audit
  const fetchReadiness = useCallback(async () => {
    if (!selectedAuditId) return;
    try {
      const [readRes, statsRes] = await Promise.all([
        getAuditReadiness(selectedAuditId).catch(() => ({ data: null })),
        getAuditStats(selectedAuditId).catch(() => ({ data: null })),
      ]);
      if (readRes.data) setReadiness(readRes.data);
      if (statsRes.data) setStats(statsRes.data);
    } catch (err) {
      console.error('Failed to fetch readiness:', err);
    }
  }, [selectedAuditId]);

  useEffect(() => { fetchReadiness(); }, [fetchReadiness]);

  if (loading) return <div className="flex items-center justify-center py-16 text-muted-foreground">Loading...</div>;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/audit">
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">Audit Readiness Dashboard</h1>
          <p className="text-sm text-muted-foreground">Track evidence request completion and identify gaps</p>
        </div>
        <div className="w-[300px]">
          <Label className="text-xs">Select Audit</Label>
          <Select value={selectedAuditId} onValueChange={setSelectedAuditId}>
            <SelectTrigger><SelectValue placeholder="Select an audit..." /></SelectTrigger>
            <SelectContent>
              {audits.map(a => (
                <SelectItem key={a.id} value={a.id}>
                  {a.title}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {!selectedAuditId ? (
        <div className="text-center py-16 text-muted-foreground">Select an audit to view readiness</div>
      ) : !readiness ? (
        <div className="text-center py-16 text-muted-foreground">Loading readiness data...</div>
      ) : (
        <>
          {/* Summary Cards */}
          <div className="grid gap-4 md:grid-cols-4">
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-3xl font-bold">{readiness.overall_readiness_pct}%</div>
                    <p className="text-xs text-muted-foreground">Overall Readiness</p>
                  </div>
                  <Target className="h-8 w-8 text-muted-foreground/40" />
                </div>
                <Progress value={readiness.overall_readiness_pct} className="mt-3" />
              </CardContent>
            </Card>
            {stats?.readiness && (
              <>
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold text-green-600">{stats.readiness.accepted}</div>
                    <p className="text-xs text-muted-foreground">Accepted Requests</p>
                    <p className="text-xs text-muted-foreground">of {stats.readiness.total_requests} total</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold text-orange-600">{stats.readiness.overdue}</div>
                    <p className="text-xs text-muted-foreground">Overdue Requests</p>
                  </CardContent>
                </Card>
              </>
            )}
            {stats?.findings && (
              <Card>
                <CardContent className="p-4">
                  <div className="text-2xl font-bold text-red-600">{(stats.findings.by_severity?.critical || 0) + (stats.findings.by_severity?.high || 0)}</div>
                  <p className="text-xs text-muted-foreground">Critical/High Findings</p>
                  <p className="text-xs text-muted-foreground">{stats.findings.total} total findings</p>
                </CardContent>
              </Card>
            )}
          </div>

          {/* By Requirement */}
          {readiness.by_requirement.length > 0 && (
            <Card>
              <CardHeader><CardTitle className="text-base">Readiness by Requirement</CardTitle></CardHeader>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Requirement</TableHead>
                      <TableHead>Total Requests</TableHead>
                      <TableHead>Accepted</TableHead>
                      <TableHead>Readiness</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {readiness.by_requirement.map((r) => (
                      <TableRow key={r.requirement_id}>
                        <TableCell className="font-medium text-sm">{r.requirement_title}</TableCell>
                        <TableCell>{r.total_requests}</TableCell>
                        <TableCell>{r.accepted_requests}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2 min-w-[140px]">
                            <Progress value={r.readiness_pct} className="flex-1" />
                            <span className="text-xs font-medium w-10 text-right">{r.readiness_pct}%</span>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {/* By Control */}
          {readiness.by_control.length > 0 && (
            <Card>
              <CardHeader><CardTitle className="text-base">Readiness by Control</CardTitle></CardHeader>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Control</TableHead>
                      <TableHead>Total Requests</TableHead>
                      <TableHead>Accepted</TableHead>
                      <TableHead>Readiness</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {readiness.by_control.map((c) => (
                      <TableRow key={c.control_id}>
                        <TableCell className="font-medium text-sm">{c.control_title}</TableCell>
                        <TableCell>{c.total_requests}</TableCell>
                        <TableCell>{c.accepted_requests}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2 min-w-[140px]">
                            <Progress value={c.readiness_pct} className="flex-1" />
                            <span className="text-xs font-medium w-10 text-right">{c.readiness_pct}%</span>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {/* Gaps */}
          {readiness.gaps.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-orange-500" />
                  Coverage Gaps ({readiness.gaps.length})
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Requirement</TableHead>
                      <TableHead>Issue</TableHead>
                      <TableHead>Description</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {readiness.gaps.map((g, i) => (
                      <TableRow key={i}>
                        <TableCell className="font-medium text-sm">{g.requirement_title}</TableCell>
                        <TableCell>
                          <Badge variant="destructive" className="text-xs">{g.issue.replace(/_/g, ' ')}</Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">{g.description}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {/* Timeline */}
          {stats?.timeline && (
            <Card>
              <CardHeader><CardTitle className="text-base">Audit Timeline</CardTitle></CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground">Planned Start:</span>
                    <p className="font-medium">{stats.timeline.planned_start ? new Date(stats.timeline.planned_start).toLocaleDateString() : '—'}</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Planned End:</span>
                    <p className="font-medium">{stats.timeline.planned_end ? new Date(stats.timeline.planned_end).toLocaleDateString() : '—'}</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Days Elapsed:</span>
                    <p className="font-medium">{stats.timeline.days_elapsed}</p>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Days Remaining:</span>
                    <p className="font-medium">{stats.timeline.days_remaining ?? '—'}</p>
                  </div>
                </div>
                {stats.timeline.next_milestone && (
                  <div className="mt-3 p-3 bg-muted rounded-lg flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm">
                      Next milestone: <strong>{stats.timeline.next_milestone.name}</strong> in {stats.timeline.next_milestone.days_until} days
                    </span>
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </>
      )}
    </div>
  );
}
