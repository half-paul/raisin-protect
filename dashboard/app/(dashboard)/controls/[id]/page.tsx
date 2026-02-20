'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  ArrowLeft,
  Shield,
  User,
  FileText,
  Link2,
  CheckCircle2,
  Clock,
  AlertTriangle,
  Archive,
  Layers,
} from 'lucide-react';
import { ControlDetail, getControl } from '@/lib/api';

const STATUS_CONFIG: Record<string, { label: string; icon: typeof CheckCircle2; color: string }> = {
  draft: { label: 'Draft', icon: Clock, color: 'text-amber-600 dark:text-amber-400' },
  active: { label: 'Active', icon: CheckCircle2, color: 'text-green-600 dark:text-green-400' },
  under_review: { label: 'Under Review', icon: AlertTriangle, color: 'text-blue-600 dark:text-blue-400' },
  deprecated: { label: 'Deprecated', icon: Archive, color: 'text-red-600 dark:text-red-400' },
};

const CATEGORY_LABELS: Record<string, string> = {
  technical: 'Technical',
  administrative: 'Administrative',
  physical: 'Physical',
  operational: 'Operational',
};

const STRENGTH_LABELS: Record<string, string> = {
  primary: 'Primary',
  supporting: 'Supporting',
  partial: 'Partial',
};

const STRENGTH_COLORS: Record<string, string> = {
  primary: 'bg-green-500/10 text-green-700 dark:text-green-400',
  supporting: 'bg-blue-500/10 text-blue-700 dark:text-blue-400',
  partial: 'bg-amber-500/10 text-amber-700 dark:text-amber-400',
};

export default function ControlDetailPage() {
  const params = useParams();
  const router = useRouter();
  const controlId = params.id as string;

  const [control, setControl] = useState<ControlDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');

  const fetchControl = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getControl(controlId);
      setControl(res.data);
    } catch {
      router.push('/controls');
    } finally {
      setLoading(false);
    }
  }, [controlId, router]);

  useEffect(() => {
    fetchControl();
  }, [fetchControl]);

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center py-24">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  if (!control) return null;

  const statusCfg = STATUS_CONFIG[control.status] || STATUS_CONFIG.draft;
  const StatusIcon = statusCfg.icon;

  // Group mappings by framework
  const mappingsByFramework: Record<string, typeof control.mappings> = {};
  control.mappings?.forEach((m) => {
    const fwName = m.requirement.framework?.name || 'Unknown';
    if (!mappingsByFramework[fwName]) mappingsByFramework[fwName] = [];
    mappingsByFramework[fwName].push(m);
  });

  return (
    <div className="p-6 space-y-6">
      {/* Back + Header */}
      <div>
        <Button variant="ghost" size="sm" className="mb-2" onClick={() => router.push('/controls')}>
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Controls
        </Button>
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-3xl font-bold tracking-tight">{control.identifier}</h1>
              <Badge variant="outline" className="text-sm">
                {CATEGORY_LABELS[control.category] || control.category}
              </Badge>
              {control.is_custom && (
                <Badge variant="secondary" className="text-sm">Custom</Badge>
              )}
            </div>
            <p className="text-lg text-muted-foreground mt-1">{control.title}</p>
          </div>
          <div className={`flex items-center gap-2 ${statusCfg.color}`}>
            <StatusIcon className="h-5 w-5" />
            <span className="font-semibold">{statusCfg.label}</span>
          </div>
        </div>
      </div>

      {/* Quick stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6 flex items-center gap-3">
            <Link2 className="h-5 w-5 text-muted-foreground" />
            <div>
              <div className="text-2xl font-bold">{control.mappings?.length || 0}</div>
              <p className="text-xs text-muted-foreground">Requirement Mappings</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6 flex items-center gap-3">
            <Layers className="h-5 w-5 text-muted-foreground" />
            <div>
              <div className="text-2xl font-bold">{Object.keys(mappingsByFramework).length}</div>
              <p className="text-xs text-muted-foreground">Frameworks</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6 flex items-center gap-3">
            <User className="h-5 w-5 text-muted-foreground" />
            <div>
              <div className="text-sm font-medium">{control.owner?.name || 'Unassigned'}</div>
              <p className="text-xs text-muted-foreground">Primary Owner</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6 flex items-center gap-3">
            <User className="h-5 w-5 text-muted-foreground" />
            <div>
              <div className="text-sm font-medium">{control.secondary_owner?.name || 'None'}</div>
              <p className="text-xs text-muted-foreground">Secondary Owner</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="overview">
            <FileText className="h-4 w-4 mr-1" />
            Overview
          </TabsTrigger>
          <TabsTrigger value="mappings">
            <Link2 className="h-4 w-4 mr-1" />
            Mappings ({control.mappings?.length || 0})
          </TabsTrigger>
        </TabsList>

        {/* Overview */}
        <TabsContent value="overview" className="space-y-6 mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Description</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm whitespace-pre-wrap">{control.description}</p>
            </CardContent>
          </Card>

          {control.implementation_guidance && (
            <Card>
              <CardHeader>
                <CardTitle>Implementation Guidance</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm whitespace-pre-wrap">{control.implementation_guidance}</p>
              </CardContent>
            </Card>
          )}

          <div className="grid gap-4 md:grid-cols-2">
            {control.evidence_requirements && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Evidence Requirements</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm whitespace-pre-wrap">{control.evidence_requirements}</p>
                </CardContent>
              </Card>
            )}
            {control.test_criteria && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Test Criteria</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm whitespace-pre-wrap">{control.test_criteria}</p>
                </CardContent>
              </Card>
            )}
          </div>

          {/* Metadata */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="text-muted-foreground">Created</p>
                  <p className="font-medium">{new Date(control.created_at).toLocaleDateString()}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Updated</p>
                  <p className="font-medium">{new Date(control.updated_at).toLocaleDateString()}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Source</p>
                  <p className="font-medium">{control.is_custom ? 'Custom' : 'Library Template'}</p>
                </div>
                {control.source_template_id && (
                  <div>
                    <p className="text-muted-foreground">Template ID</p>
                    <p className="font-mono text-xs">{control.source_template_id}</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Mappings */}
        <TabsContent value="mappings" className="space-y-6 mt-4">
          {Object.keys(mappingsByFramework).length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <Link2 className="h-12 w-12 mb-4 opacity-50" />
                <p className="text-lg font-medium">No mappings</p>
                <p className="text-sm">This control is not mapped to any requirements yet</p>
              </CardContent>
            </Card>
          ) : (
            Object.entries(mappingsByFramework).map(([fwName, mappings]) => (
              <Card key={fwName}>
                <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                    <Shield className="h-4 w-4" />
                    {fwName}
                  </CardTitle>
                  <CardDescription>
                    {mappings.length} requirement{mappings.length !== 1 ? 's' : ''} mapped
                  </CardDescription>
                </CardHeader>
                <CardContent className="p-0">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[120px]">Requirement</TableHead>
                        <TableHead>Title</TableHead>
                        <TableHead className="w-[100px]">Strength</TableHead>
                        <TableHead>Notes</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {mappings.map((m) => (
                        <TableRow key={m.id}>
                          <TableCell className="font-mono text-sm font-medium">
                            {m.requirement.identifier}
                          </TableCell>
                          <TableCell className="max-w-sm">
                            <span className="line-clamp-1 text-sm">{m.requirement.title}</span>
                          </TableCell>
                          <TableCell>
                            <span className={`text-xs font-medium px-2 py-1 rounded-full ${STRENGTH_COLORS[m.strength] || ''}`}>
                              {STRENGTH_LABELS[m.strength] || m.strength}
                            </span>
                          </TableCell>
                          <TableCell className="max-w-xs">
                            <span className="line-clamp-1 text-xs text-muted-foreground">
                              {m.notes || '—'}
                            </span>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
            ))
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
