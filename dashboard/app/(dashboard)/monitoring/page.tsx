'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Tooltip, TooltipContent, TooltipProvider, TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  Activity, AlertTriangle, CheckCircle2, XCircle, HelpCircle,
  Shield, TrendingUp, TrendingDown, Minus, Bell, Clock,
  RefreshCw, ArrowRight,
} from 'lucide-react';
import {
  MonitoringSummary, HeatmapData, HeatmapControl, PostureData,
  getMonitoringSummary, getMonitoringHeatmap, getMonitoringPosture,
} from '@/lib/api';
import { cn } from '@/lib/utils';

const HEALTH_COLORS: Record<string, string> = {
  healthy: 'bg-green-500',
  failing: 'bg-red-500',
  error: 'bg-orange-500',
  warning: 'bg-amber-500',
  untested: 'bg-gray-300 dark:bg-gray-600',
};

const HEALTH_TEXT_COLORS: Record<string, string> = {
  healthy: 'text-green-600 dark:text-green-400',
  failing: 'text-red-600 dark:text-red-400',
  error: 'text-orange-600 dark:text-orange-400',
  warning: 'text-amber-600 dark:text-amber-400',
  untested: 'text-gray-500 dark:text-gray-400',
};

const HEALTH_LABELS: Record<string, string> = {
  healthy: 'Healthy',
  failing: 'Failing',
  error: 'Error',
  warning: 'Warning',
  untested: 'Untested',
};

function formatRelativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  const days = Math.floor(hrs / 24);
  return `${days}d ago`;
}

function HeatmapCell({ control }: { control: HeatmapControl }) {
  return (
    <TooltipProvider delayDuration={200}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Link href={`/controls/${control.id}`}>
            <div
              className={cn(
                'w-8 h-8 rounded-sm cursor-pointer transition-all hover:ring-2 hover:ring-ring hover:ring-offset-1',
                HEALTH_COLORS[control.health_status]
              )}
            />
          </Link>
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-xs">
          <div className="space-y-1">
            <p className="font-mono text-xs">{control.identifier}</p>
            <p className="font-medium text-sm">{control.title}</p>
            <div className="flex items-center gap-2">
              <Badge variant="outline" className="text-[10px]">
                {HEALTH_LABELS[control.health_status]}
              </Badge>
              {control.tests_count > 0 && (
                <span className="text-xs text-muted-foreground">
                  {control.tests_count} test{control.tests_count !== 1 ? 's' : ''}
                </span>
              )}
              {control.active_alerts > 0 && (
                <span className="text-xs text-red-600 dark:text-red-400">
                  {control.active_alerts} alert{control.active_alerts !== 1 ? 's' : ''}
                </span>
              )}
            </div>
            {control.latest_result?.tested_at && (
              <p className="text-xs text-muted-foreground">
                Last tested: {formatRelativeTime(control.latest_result.tested_at)}
              </p>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

export default function MonitoringDashboardPage() {
  const [summary, setSummary] = useState<MonitoringSummary | null>(null);
  const [heatmap, setHeatmap] = useState<HeatmapData | null>(null);
  const [posture, setPosture] = useState<PostureData | null>(null);
  const [loading, setLoading] = useState(true);
  const [categoryFilter, setCategoryFilter] = useState('');

  async function fetchData() {
    setLoading(true);
    try {
      const heatmapParams: Record<string, string> = {};
      if (categoryFilter) heatmapParams.category = categoryFilter;

      const [summaryRes, heatmapRes, postureRes] = await Promise.all([
        getMonitoringSummary(),
        getMonitoringHeatmap(heatmapParams),
        getMonitoringPosture(),
      ]);
      setSummary(summaryRes.data);
      setHeatmap(heatmapRes.data);
      setPosture(postureRes.data);
    } catch {
      // Dashboard handles missing data gracefully
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { fetchData(); }, [categoryFilter]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
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
            <Activity className="h-8 w-8" />
            Monitoring Dashboard
          </h1>
          <p className="text-muted-foreground mt-1">
            Real-time compliance posture and control health
          </p>
        </div>
        <Button variant="outline" onClick={() => fetchData()}>
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Top-level stats */}
      {summary && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Posture Score</CardTitle>
              <Shield className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">
                {summary.overall_posture_score.toFixed(1)}%
              </div>
              <p className="text-xs text-muted-foreground">
                {summary.controls.health_rate.toFixed(0)}% controls healthy
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Test Pass Rate (24h)</CardTitle>
              <CheckCircle2 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-green-600 dark:text-green-400">
                {summary.tests.pass_rate_24h.toFixed(1)}%
              </div>
              <p className="text-xs text-muted-foreground">
                {summary.tests.total_active} active tests
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Open Alerts</CardTitle>
              <Bell className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-amber-600 dark:text-amber-400">
                {summary.alerts.open + summary.alerts.acknowledged + summary.alerts.in_progress}
              </div>
              <div className="flex items-center gap-2 mt-1">
                {summary.alerts.by_severity.critical > 0 && (
                  <Badge variant="destructive" className="text-[10px]">
                    {summary.alerts.by_severity.critical} critical
                  </Badge>
                )}
                {summary.alerts.by_severity.high > 0 && (
                  <Badge variant="outline" className="text-[10px] text-orange-600 border-orange-600/20">
                    {summary.alerts.by_severity.high} high
                  </Badge>
                )}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">SLA Breached</CardTitle>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className={cn(
                'text-3xl font-bold',
                summary.alerts.sla_breached > 0
                  ? 'text-red-600 dark:text-red-400'
                  : 'text-green-600 dark:text-green-400'
              )}>
                {summary.alerts.sla_breached}
              </div>
              <p className="text-xs text-muted-foreground">
                {summary.alerts.resolved_today} resolved today
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Control Health Heatmap */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-lg flex items-center gap-2">
                Control Health Heatmap
              </CardTitle>
              <CardDescription>
                Each cell represents a control — color indicates health status
              </CardDescription>
            </div>
            <Select value={categoryFilter} onValueChange={(v) => setCategoryFilter(v === 'all' ? '' : v)}>
              <SelectTrigger className="w-[160px]">
                <SelectValue placeholder="All Categories" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Categories</SelectItem>
                <SelectItem value="technical">Technical</SelectItem>
                <SelectItem value="administrative">Administrative</SelectItem>
                <SelectItem value="physical">Physical</SelectItem>
                <SelectItem value="operational">Operational</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
        <CardContent>
          {/* Legend */}
          <div className="flex items-center gap-4 mb-4">
            {Object.entries(HEALTH_LABELS).map(([key, label]) => (
              <div key={key} className="flex items-center gap-1.5">
                <div className={cn('w-3 h-3 rounded-sm', HEALTH_COLORS[key])} />
                <span className="text-xs text-muted-foreground">{label}</span>
              </div>
            ))}
          </div>

          {/* Summary bar */}
          {heatmap?.summary && (
            <div className="flex items-center gap-4 mb-4 text-sm">
              <span className={HEALTH_TEXT_COLORS.healthy}>
                <CheckCircle2 className="h-4 w-4 inline mr-1" />
                {heatmap.summary.healthy} healthy
              </span>
              <span className={HEALTH_TEXT_COLORS.failing}>
                <XCircle className="h-4 w-4 inline mr-1" />
                {heatmap.summary.failing} failing
              </span>
              {heatmap.summary.error > 0 && (
                <span className={HEALTH_TEXT_COLORS.error}>
                  <AlertTriangle className="h-4 w-4 inline mr-1" />
                  {heatmap.summary.error} error
                </span>
              )}
              {heatmap.summary.warning > 0 && (
                <span className={HEALTH_TEXT_COLORS.warning}>
                  <AlertTriangle className="h-4 w-4 inline mr-1" />
                  {heatmap.summary.warning} warning
                </span>
              )}
              <span className={HEALTH_TEXT_COLORS.untested}>
                <HelpCircle className="h-4 w-4 inline mr-1" />
                {heatmap.summary.untested} untested
              </span>
              <span className="text-muted-foreground ml-auto">
                {heatmap.summary.total_controls} total controls
              </span>
            </div>
          )}

          {/* Heatmap grid */}
          {heatmap?.controls && heatmap.controls.length > 0 ? (
            <div className="flex flex-wrap gap-1">
              {heatmap.controls.map((ctrl) => (
                <HeatmapCell key={ctrl.id} control={ctrl} />
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <HelpCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No control health data available</p>
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2">
        {/* Compliance Posture */}
        {posture && (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Shield className="h-5 w-5" />
                Compliance Posture
              </CardTitle>
              <CardDescription>Score per activated framework</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold mb-4">
                {posture.overall_score.toFixed(1)}%
                <span className="text-sm font-normal text-muted-foreground ml-2">overall</span>
              </div>
              <div className="space-y-4">
                {posture.frameworks.map((fw) => (
                  <div key={fw.framework_id} className="space-y-1.5">
                    <div className="flex items-center justify-between text-sm">
                      <span className="font-medium">
                        {fw.framework_name} {fw.framework_version}
                      </span>
                      <div className="flex items-center gap-2">
                        <span className="font-bold">{fw.posture_score.toFixed(1)}%</span>
                        {fw.trend && (
                          <span className={cn(
                            'flex items-center text-xs',
                            fw.trend.direction === 'improving' ? 'text-green-600 dark:text-green-400' :
                            fw.trend.direction === 'declining' ? 'text-red-600 dark:text-red-400' :
                            'text-muted-foreground'
                          )}>
                            {fw.trend.direction === 'improving' && <TrendingUp className="h-3 w-3 mr-0.5" />}
                            {fw.trend.direction === 'declining' && <TrendingDown className="h-3 w-3 mr-0.5" />}
                            {fw.trend.direction === 'stable' && <Minus className="h-3 w-3 mr-0.5" />}
                            {fw.trend.direction}
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="h-2 rounded-full bg-secondary">
                      <div
                        className={cn(
                          'h-full rounded-full transition-all',
                          fw.posture_score >= 80 ? 'bg-green-500' :
                          fw.posture_score >= 60 ? 'bg-amber-500' :
                          'bg-red-500'
                        )}
                        style={{ width: `${fw.posture_score}%` }}
                      />
                    </div>
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>{fw.passing} passing</span>
                      <span>{fw.failing} failing</span>
                      <span>{fw.untested} untested</span>
                    </div>
                  </div>
                ))}
                {posture.frameworks.length === 0 && (
                  <p className="text-sm text-muted-foreground">No frameworks activated</p>
                )}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Recent Activity & Quick Links */}
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Activity className="h-5 w-5" />
                Recent Activity
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {summary?.recent_activity && summary.recent_activity.length > 0 ? (
                  summary.recent_activity.slice(0, 8).map((event, i) => (
                    <div key={i} className="flex items-start justify-between text-sm border-b border-border/50 pb-2 last:border-0">
                      <div className="flex items-center gap-2">
                        {event.type === 'alert_created' && <AlertTriangle className="h-4 w-4 text-amber-500 shrink-0" />}
                        {event.type === 'alert_resolved' && <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0" />}
                        {event.type === 'test_run_completed' && <Activity className="h-4 w-4 text-blue-500 shrink-0" />}
                        <div>
                          {event.type === 'alert_created' && (
                            <span>Alert #{event.alert_number}: <span className="text-muted-foreground">{event.title}</span></span>
                          )}
                          {event.type === 'alert_resolved' && (
                            <span>Resolved #{event.alert_number} <span className="text-muted-foreground">by {event.resolved_by}</span></span>
                          )}
                          {event.type === 'test_run_completed' && (
                            <span>Run #{event.run_number}: <span className="text-green-600 dark:text-green-400">{event.passed} passed</span>, <span className="text-red-600 dark:text-red-400">{event.failed} failed</span></span>
                          )}
                          {!['alert_created', 'alert_resolved', 'test_run_completed'].includes(event.type) && (
                            <span>{event.title || event.type}</span>
                          )}
                        </div>
                      </div>
                      <span className="text-xs text-muted-foreground whitespace-nowrap ml-2">
                        {formatRelativeTime(event.timestamp)}
                      </span>
                    </div>
                  ))
                ) : (
                  <p className="text-sm text-muted-foreground">No recent activity</p>
                )}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Quick Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Link href="/alerts">
                <Button variant="outline" className="w-full justify-between">
                  View Alert Queue
                  <ArrowRight className="h-4 w-4" />
                </Button>
              </Link>
              <Link href="/test-runs">
                <Button variant="outline" className="w-full justify-between">
                  Test Execution History
                  <ArrowRight className="h-4 w-4" />
                </Button>
              </Link>
              <Link href="/alert-rules">
                <Button variant="outline" className="w-full justify-between">
                  Manage Alert Rules
                  <ArrowRight className="h-4 w-4" />
                </Button>
              </Link>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
