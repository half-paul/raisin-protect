'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { getRoleLabel } from '@/lib/auth';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Shield,
  AlertTriangle,
  FileCheck,
  Users,
  ClipboardCheck,
  Truck,
  Monitor,
  TrendingUp,
} from 'lucide-react';
import { ControlStats, OrgFramework, getControlStats, listOrgFrameworks } from '@/lib/api';

interface StatCardProps {
  title: string;
  value: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  trend?: string;
}

function StatCard({ title, value, description, icon: Icon, trend }: StatCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <p className="text-xs text-muted-foreground">
          {trend && <span className="text-green-600 dark:text-green-400">{trend} </span>}
          {description}
        </p>
      </CardContent>
    </Card>
  );
}

export default function DashboardPage() {
  const { user } = useAuth();
  const [stats, setStats] = useState<ControlStats | null>(null);
  const [orgFrameworks, setOrgFrameworks] = useState<OrgFramework[]>([]);

  useEffect(() => {
    async function fetchDashData() {
      try {
        const [statsRes, ofRes] = await Promise.all([
          getControlStats(),
          listOrgFrameworks({ status: 'active' }),
        ]);
        setStats(statsRes.data);
        setOrgFrameworks(ofRes.data || []);
      } catch {
        // Dashboard loads with static defaults if API unavailable
      }
    }
    fetchDashData();
  }, []);

  const overallCoverage = stats?.frameworks_coverage?.length
    ? (stats.frameworks_coverage.reduce((s, f) => s + f.covered, 0) /
        Math.max(stats.frameworks_coverage.reduce((s, f) => s + f.in_scope, 0), 1)) * 100
    : 0;

  return (
    <div className="p-6 space-y-6">
      {/* Welcome header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          Welcome back{user?.first_name ? `, ${user.first_name}` : ''}
        </h1>
        <p className="text-muted-foreground mt-1">
          {user?.role ? getRoleLabel(user.role) : 'GRC Dashboard'} — Here&apos;s your compliance overview.
        </p>
      </div>

      {/* Stats grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Compliance Score"
          value={`${overallCoverage.toFixed(0)}%`}
          description="across all frameworks"
          icon={Shield}
        />
        <StatCard
          title="Active Frameworks"
          value={String(orgFrameworks.length)}
          description="compliance frameworks tracked"
          icon={FileCheck}
        />
        <StatCard
          title="Controls"
          value={String(stats?.total || 0)}
          description={`${stats?.by_status?.active || 0} active, ${stats?.by_status?.draft || 0} draft`}
          icon={Shield}
        />
        <StatCard
          title="Coverage Gaps"
          value={String(stats?.unmapped_count || 0)}
          description="controls without mappings"
          icon={AlertTriangle}
        />
      </div>

      {/* Role-specific sections */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <ClipboardCheck className="h-5 w-5" />
              Recent Audit Activity
            </CardTitle>
            <CardDescription>Latest audit trail entries</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {[
                { action: 'User login', actor: 'alice@acme.com', time: '2 min ago' },
                { action: 'Control updated', actor: 'bob@acme.com', time: '15 min ago' },
                { action: 'Evidence uploaded', actor: 'carol@acme.com', time: '1 hour ago' },
                { action: 'Role changed', actor: 'dave@acme.com', time: '3 hours ago' },
              ].map((entry, i) => (
                <div key={i} className="flex items-center justify-between text-sm">
                  <div>
                    <span className="font-medium">{entry.action}</span>
                    <span className="text-muted-foreground"> by {entry.actor}</span>
                  </div>
                  <span className="text-xs text-muted-foreground">{entry.time}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Monitor className="h-5 w-5" />
              System Health
            </CardTitle>
            <CardDescription>Service and integration status</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {[
                { name: 'API Server', status: 'healthy' },
                { name: 'Database', status: 'healthy' },
                { name: 'Redis Cache', status: 'healthy' },
                { name: 'Email Service', status: 'pending' },
              ].map((service, i) => (
                <div key={i} className="flex items-center justify-between text-sm">
                  <span>{service.name}</span>
                  <Badge variant={service.status === 'healthy' ? 'default' : 'secondary'}>
                    {service.status}
                  </Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <TrendingUp className="h-5 w-5" />
              Framework Coverage
            </CardTitle>
            <CardDescription>Compliance by framework</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {orgFrameworks.length > 0 ? (
                orgFrameworks.map((of) => (
                  <Link key={of.id} href={`/frameworks/${of.id}`} className="block">
                    <div className="space-y-1">
                      <div className="flex items-center justify-between text-sm">
                        <span>{of.framework.name}</span>
                        <span className="font-medium">{of.stats.coverage_pct.toFixed(0)}%</span>
                      </div>
                      <div className="h-2 rounded-full bg-secondary">
                        <div
                          className="h-full rounded-full bg-primary transition-all"
                          style={{ width: `${of.stats.coverage_pct}%` }}
                        />
                      </div>
                    </div>
                  </Link>
                ))
              ) : (
                <p className="text-sm text-muted-foreground">No frameworks activated yet</p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
