'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Plug,
  Search,
  Plus,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  XCircle,
  Clock,
  Activity,
  Wifi,
  WifiOff,
  Loader2,
} from 'lucide-react';
import {
  IntegrationDefinition,
  IntegrationConnection,
  IntegrationDashboardData,
  listIntegrations,
  listIntegrationConnections,
  getIntegrationDashboard,
  createIntegrationConnection,
  testIntegrationConnection,
  enableIntegrationConnection,
  deleteIntegrationConnection,
  triggerIntegrationSync,
} from '@/lib/api';
import { WikiHelpLink } from '@/components/wiki-help-link';

const STATUS_COLORS: Record<string, string> = {
  connected: 'bg-green-100 text-green-800',
  pending: 'bg-yellow-100 text-yellow-800',
  disconnected: 'bg-gray-100 text-gray-800',
  error: 'bg-red-100 text-red-800',
  disabled: 'bg-gray-100 text-gray-500',
};

const HEALTH_COLORS: Record<string, string> = {
  healthy: 'bg-green-100 text-green-800',
  degraded: 'bg-yellow-100 text-yellow-800',
  unhealthy: 'bg-red-100 text-red-800',
  unknown: 'bg-gray-100 text-gray-500',
};

const CATEGORY_LABELS: Record<string, string> = {
  cloud_infrastructure: 'Cloud Infrastructure',
  identity_provider: 'Identity Provider',
  version_control: 'Version Control',
  communication: 'Communication',
  monitoring: 'Monitoring',
  ticketing: 'Ticketing',
  custom: 'Custom',
};

export default function IntegrationsPage() {
  const { hasRole } = useAuth();
  const canManage = hasRole('ciso', 'compliance_manager', 'it_admin', 'devops_engineer');

  const [definitions, setDefinitions] = useState<IntegrationDefinition[]>([]);
  const [connections, setConnections] = useState<IntegrationConnection[]>([]);
  const [dashboard, setDashboard] = useState<IntegrationDashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters (debounce search to avoid API call per keystroke)
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [categoryFilter, setCategoryFilter] = useState('all');

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(timer);
  }, [search]);

  // Add connection dialog
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [selectedDef, setSelectedDef] = useState<IntegrationDefinition | null>(null);
  const [newConnName, setNewConnName] = useState('');
  const [newConnDesc, setNewConnDesc] = useState('');
  const [creating, setCreating] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {};
      if (statusFilter !== 'all') params.status = statusFilter;
      if (categoryFilter !== 'all') params.category = categoryFilter;

      const [defsRes, connsRes, dashRes] = await Promise.all([
        listIntegrations({ search: debouncedSearch }),
        listIntegrationConnections(params),
        getIntegrationDashboard(),
      ]);
      setDefinitions(defsRes.data || []);
      setConnections(connsRes.data || []);
      setDashboard(dashRes.data || null);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load integrations');
    } finally {
      setLoading(false);
    }
  }, [debouncedSearch, statusFilter, categoryFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCreateConnection = async () => {
    if (!selectedDef || !newConnName.trim()) return;
    try {
      setCreating(true);
      await createIntegrationConnection({
        definition_id: selectedDef.id,
        name: newConnName.trim(),
        description: newConnDesc.trim() || undefined,
      });
      setShowAddDialog(false);
      setSelectedDef(null);
      setNewConnName('');
      setNewConnDesc('');
      fetchData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create connection');
    } finally {
      setCreating(false);
    }
  };

  const handleTestConnection = async (id: string) => {
    try {
      await testIntegrationConnection(id);
      fetchData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Test failed');
    }
  };

  const handleEnableConnection = async (id: string) => {
    try {
      await enableIntegrationConnection(id);
      fetchData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to enable');
    }
  };

  const handleSync = async (id: string) => {
    try {
      await triggerIntegrationSync(id);
      fetchData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sync failed');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this connection?')) return;
    try {
      await deleteIntegrationConnection(id);
      fetchData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Delete failed');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold">Integrations</h1>
            <WikiHelpLink path="integrations/" label="Integration Engine docs" />
          </div>
          <p className="text-muted-foreground">Connect third-party services for automated evidence collection and monitoring</p>
        </div>
        {canManage && (
          <Button onClick={() => setShowAddDialog(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Add Integration
          </Button>
        )}
      </div>

      {/* Dashboard summary cards */}
      {dashboard && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Connections</CardDescription>
              <CardTitle className="text-2xl">{dashboard.connections.total}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex gap-2 text-xs">
                <span className="text-green-600">{dashboard.connections.connected} active</span>
                {dashboard.connections.error > 0 && (
                  <span className="text-red-600">{dashboard.connections.error} error</span>
                )}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Health</CardDescription>
              <CardTitle className="text-2xl flex items-center gap-2">
                {dashboard.health.healthy}
                <Activity className="h-5 w-5 text-green-500" />
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex gap-2 text-xs">
                <span className="text-green-600">{dashboard.health.healthy} healthy</span>
                {dashboard.health.degraded > 0 && <span className="text-yellow-600">{dashboard.health.degraded} degraded</span>}
                {dashboard.health.unhealthy > 0 && <span className="text-red-600">{dashboard.health.unhealthy} unhealthy</span>}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Recent Failures (24h)</CardDescription>
              <CardTitle className="text-2xl">{dashboard.recent_failures}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-xs text-muted-foreground">
                {dashboard.active_syncs} syncs in progress
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Stale Connections</CardDescription>
              <CardTitle className="text-2xl">{dashboard.stale_connections}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-xs text-muted-foreground">
                No sync in 24+ hours
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <div className="relative flex-1 min-w-[200px] max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search integrations..."
            className="pl-9"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Statuses</SelectItem>
            <SelectItem value="connected">Connected</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="disconnected">Disconnected</SelectItem>
            <SelectItem value="error">Error</SelectItem>
            <SelectItem value="disabled">Disabled</SelectItem>
          </SelectContent>
        </Select>
        <Select value={categoryFilter} onValueChange={setCategoryFilter}>
          <SelectTrigger className="w-[200px]">
            <SelectValue placeholder="Category" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Categories</SelectItem>
            {Object.entries(CATEGORY_LABELS).map(([key, label]) => (
              <SelectItem key={key} value={key}>{label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button variant="outline" size="icon" onClick={fetchData}>
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-700">
          <AlertCircle className="inline h-4 w-4 mr-1" />
          {error}
        </div>
      )}

      {/* Connections List */}
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : connections.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Plug className="h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium">No connections yet</h3>
            <p className="text-muted-foreground text-sm mt-1">Add an integration to get started with automated evidence collection</p>
            {canManage && (
              <Button className="mt-4" onClick={() => setShowAddDialog(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Add Integration
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {connections.map((conn) => (
            <Card key={conn.id} className="hover:shadow-md transition-shadow">
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="h-10 w-10 bg-muted rounded-lg flex items-center justify-center">
                      <Plug className="h-5 w-5 text-muted-foreground" />
                    </div>
                    <div>
                      <Link href={`/integrations/${conn.id}`} className="hover:underline">
                        <CardTitle className="text-base">{conn.name}</CardTitle>
                      </Link>
                      <CardDescription>{conn.definition_name}</CardDescription>
                    </div>
                  </div>
                  <Badge className={HEALTH_COLORS[conn.health] || HEALTH_COLORS.unknown}>
                    {conn.health}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Status</span>
                    <Badge variant="outline" className={STATUS_COLORS[conn.status] || ''}>
                      {conn.status === 'connected' && <Wifi className="mr-1 h-3 w-3" />}
                      {conn.status === 'disconnected' && <WifiOff className="mr-1 h-3 w-3" />}
                      {conn.status === 'error' && <XCircle className="mr-1 h-3 w-3" />}
                      {conn.status}
                    </Badge>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Last Sync</span>
                    <span>{conn.last_sync_at ? new Date(conn.last_sync_at).toLocaleString() : 'Never'}</span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Runs</span>
                    <span>
                      {conn.successful_runs}/{conn.total_runs} successful
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Category</span>
                    <span className="text-xs">{CATEGORY_LABELS[conn.definition_category] || conn.definition_category}</span>
                  </div>
                  {canManage && (
                    <div className="flex gap-2 pt-2 border-t">
                      {conn.status === 'connected' && (
                        <Button variant="outline" size="sm" onClick={() => handleSync(conn.id)}>
                          <RefreshCw className="mr-1 h-3 w-3" /> Sync
                        </Button>
                      )}
                      {conn.status === 'pending' && (
                        <>
                          <Button variant="outline" size="sm" onClick={() => handleTestConnection(conn.id)}>
                            Test
                          </Button>
                          <Button variant="outline" size="sm" onClick={() => handleEnableConnection(conn.id)}>
                            Enable
                          </Button>
                        </>
                      )}
                      <Link href={`/integrations/${conn.id}`}>
                        <Button variant="ghost" size="sm">Details</Button>
                      </Link>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Integration Catalog (when adding) */}
      {!showAddDialog && connections.length > 0 && definitions.length > 0 && (
        <div>
          <h2 className="text-lg font-semibold mb-3">Available Integrations</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {definitions.map((def) => (
              <Card key={def.id} className="hover:shadow-sm transition-shadow cursor-pointer" onClick={() => {
                if (canManage) {
                  setSelectedDef(def);
                  setShowAddDialog(true);
                }
              }}>
                <CardHeader className="pb-2">
                  <div className="flex items-center gap-3">
                    <div className="h-8 w-8 bg-muted rounded flex items-center justify-center">
                      <Plug className="h-4 w-4 text-muted-foreground" />
                    </div>
                    <div>
                      <CardTitle className="text-sm">{def.name}</CardTitle>
                      <CardDescription className="text-xs">{def.short_description}</CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="flex flex-wrap gap-1">
                    {def.capabilities.slice(0, 3).map((cap) => (
                      <Badge key={cap} variant="secondary" className="text-xs">{cap.replace(/_/g, ' ')}</Badge>
                    ))}
                    {def.is_beta && <Badge variant="outline" className="text-xs text-orange-600">Beta</Badge>}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Add Connection Dialog */}
      <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Add Integration Connection</DialogTitle>
            <DialogDescription>
              {selectedDef ? `Configure a new ${selectedDef.name} connection` : 'Select an integration to connect'}
            </DialogDescription>
          </DialogHeader>

          {!selectedDef ? (
            <div className="grid gap-2 max-h-[400px] overflow-y-auto">
              {definitions.map((def) => (
                <button
                  key={def.id}
                  className="flex items-center gap-3 p-3 rounded-lg border hover:bg-accent text-left"
                  onClick={() => setSelectedDef(def)}
                >
                  <div className="h-10 w-10 bg-muted rounded-lg flex items-center justify-center flex-shrink-0">
                    <Plug className="h-5 w-5 text-muted-foreground" />
                  </div>
                  <div>
                    <div className="font-medium text-sm">{def.name}</div>
                    <div className="text-xs text-muted-foreground">{def.short_description}</div>
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <div className="space-y-4">
              <div className="flex items-center gap-3 p-3 bg-muted rounded-lg">
                <Plug className="h-5 w-5" />
                <div>
                  <div className="font-medium text-sm">{selectedDef.name}</div>
                  <div className="text-xs text-muted-foreground">{selectedDef.short_description}</div>
                </div>
              </div>
              <div className="space-y-2">
                <Label htmlFor="conn-name">Connection Name</Label>
                <Input
                  id="conn-name"
                  placeholder={`${selectedDef.name} Production`}
                  value={newConnName}
                  onChange={(e) => setNewConnName(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="conn-desc">Description (optional)</Label>
                <Input
                  id="conn-desc"
                  placeholder="Brief description of this connection"
                  value={newConnDesc}
                  onChange={(e) => setNewConnDesc(e.target.value)}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            {selectedDef && (
              <Button variant="outline" onClick={() => setSelectedDef(null)}>Back</Button>
            )}
            <Button variant="outline" onClick={() => { setShowAddDialog(false); setSelectedDef(null); }}>Cancel</Button>
            {selectedDef && (
              <Button onClick={handleCreateConnection} disabled={!newConnName.trim() || creating}>
                {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create Connection
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
