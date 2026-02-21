'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import { Search, BookOpen, Copy, FileText } from 'lucide-react';
import {
  PolicyTemplate, OrgFramework,
  listPolicyTemplates, clonePolicyTemplate, listOrgFrameworks,
} from '@/lib/api';

const CATEGORY_LABELS: Record<string, string> = {
  information_security: 'Information Security', access_control: 'Access Control',
  incident_response: 'Incident Response', data_privacy: 'Data Privacy',
  network_security: 'Network Security', encryption: 'Encryption',
  vulnerability_management: 'Vulnerability Management', change_management: 'Change Management',
  business_continuity: 'Business Continuity', secure_development: 'Secure Development',
  vendor_management: 'Vendor Management', acceptable_use: 'Acceptable Use',
  physical_security: 'Physical Security', hr_security: 'HR Security',
  asset_management: 'Asset Management',
};

export default function PolicyTemplateLibraryPage() {
  const router = useRouter();
  const { hasRole } = useAuth();
  const canClone = hasRole('ciso', 'compliance_manager', 'security_engineer');

  const [templates, setTemplates] = useState<PolicyTemplate[]>([]);
  const [frameworks, setFrameworks] = useState<OrgFramework[]>([]);
  const [loading, setLoading] = useState(true);

  // Filters
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [frameworkFilter, setFrameworkFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');

  // Clone dialog
  const [cloneTemplate, setCloneTemplate] = useState<PolicyTemplate | null>(null);
  const [cloneForm, setCloneForm] = useState({
    identifier: '',
    title: '',
    description: '',
    review_frequency_days: '',
    tags: '',
  });
  const [cloning, setCloning] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {};
      if (search) params.search = search;
      if (frameworkFilter) params.framework_id = frameworkFilter;
      if (categoryFilter) params.category = categoryFilter;

      const [tplRes, fwRes] = await Promise.all([
        listPolicyTemplates(params),
        listOrgFrameworks(),
      ]);
      setTemplates(tplRes.data);
      setFrameworks(fwRes.data);
    } catch (err) {
      console.error('Failed to fetch templates:', err);
    } finally {
      setLoading(false);
    }
  }, [search, frameworkFilter, categoryFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const openClone = (tpl: PolicyTemplate) => {
    setCloneTemplate(tpl);
    setCloneForm({
      identifier: tpl.identifier.replace('TPL-', 'POL-'),
      title: tpl.title,
      description: tpl.description || '',
      review_frequency_days: String(tpl.review_frequency_days || 365),
      tags: (tpl.tags || []).filter(t => t !== 'template').join(', '),
    });
  };

  const handleClone = async () => {
    if (!cloneTemplate || !cloneForm.identifier) return;
    try {
      setCloning(true);
      const res = await clonePolicyTemplate(cloneTemplate.id, {
        identifier: cloneForm.identifier,
        title: cloneForm.title || undefined,
        description: cloneForm.description || undefined,
        review_frequency_days: cloneForm.review_frequency_days ? parseInt(cloneForm.review_frequency_days) : undefined,
        tags: cloneForm.tags ? cloneForm.tags.split(',').map(t => t.trim()).filter(Boolean) : undefined,
      });
      setCloneTemplate(null);
      router.push(`/policies/${res.data.id}`);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to clone template');
    } finally {
      setCloning(false);
    }
  };

  // Group templates by framework
  const groupedTemplates: Record<string, PolicyTemplate[]> = {};
  templates.forEach(tpl => {
    const key = tpl.framework?.name || 'General';
    if (!groupedTemplates[key]) groupedTemplates[key] = [];
    groupedTemplates[key].push(tpl);
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <BookOpen className="h-6 w-6" /> Policy Template Library
        </h1>
        <p className="text-sm text-muted-foreground">Browse and clone pre-built policy templates for your organization</p>
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
                  placeholder="Search templates..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') setSearch(searchInput); }}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="w-[180px]">
              <Label className="text-xs">Framework</Label>
              <Select value={frameworkFilter} onValueChange={(v) => setFrameworkFilter(v === 'all' ? '' : v)}>
                <SelectTrigger><SelectValue placeholder="All frameworks" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All frameworks</SelectItem>
                  {frameworks.map(fw => (
                    <SelectItem key={fw.id} value={fw.framework.id}>{fw.framework.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[180px]">
              <Label className="text-xs">Category</Label>
              <Select value={categoryFilter} onValueChange={(v) => setCategoryFilter(v === 'all' ? '' : v)}>
                <SelectTrigger><SelectValue placeholder="All categories" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All categories</SelectItem>
                  {Object.entries(CATEGORY_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Templates */}
      {loading ? (
        <div className="text-center py-12 text-muted-foreground">Loading templates...</div>
      ) : templates.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">No templates found</div>
      ) : (
        Object.entries(groupedTemplates).map(([framework, tpls]) => (
          <div key={framework}>
            <h2 className="text-lg font-semibold mb-3">{framework}</h2>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {tpls.map((tpl) => (
                <Card key={tpl.id} className="hover:shadow-md transition-shadow">
                  <CardHeader className="pb-2">
                    <div className="flex items-start justify-between">
                      <div>
                        <CardTitle className="text-base">{tpl.title}</CardTitle>
                        <CardDescription className="text-xs mt-1">{tpl.identifier}</CardDescription>
                      </div>
                      <FileText className="h-5 w-5 text-muted-foreground flex-shrink-0" />
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {tpl.description || 'No description available.'}
                    </p>
                    <div className="flex flex-wrap gap-1">
                      <Badge variant="outline" className="text-xs">{CATEGORY_LABELS[tpl.category] || tpl.category}</Badge>
                      {tpl.current_version?.word_count && (
                        <Badge variant="secondary" className="text-xs">{tpl.current_version.word_count} words</Badge>
                      )}
                      {tpl.review_frequency_days && (
                        <Badge variant="secondary" className="text-xs">Review: {tpl.review_frequency_days}d</Badge>
                      )}
                    </div>
                    {tpl.tags && tpl.tags.length > 0 && (
                      <div className="flex flex-wrap gap-1">
                        {tpl.tags.filter(t => t !== 'template').slice(0, 5).map(tag => (
                          <span key={tag} className="text-xs text-muted-foreground bg-muted rounded px-1.5 py-0.5">{tag}</span>
                        ))}
                      </div>
                    )}
                    {canClone && (
                      <Button variant="outline" size="sm" className="w-full" onClick={() => openClone(tpl)}>
                        <Copy className="h-4 w-4 mr-2" /> Clone to My Policies
                      </Button>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        ))
      )}

      {/* Clone Dialog */}
      <Dialog open={!!cloneTemplate} onOpenChange={() => setCloneTemplate(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Clone Template</DialogTitle>
            <DialogDescription>
              Create a new policy from &ldquo;{cloneTemplate?.title}&rdquo;
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Policy Identifier *</Label>
              <Input value={cloneForm.identifier} onChange={(e) => setCloneForm(f => ({ ...f, identifier: e.target.value }))} />
            </div>
            <div>
              <Label>Title</Label>
              <Input value={cloneForm.title} onChange={(e) => setCloneForm(f => ({ ...f, title: e.target.value }))} />
            </div>
            <div>
              <Label>Description</Label>
              <Input value={cloneForm.description} onChange={(e) => setCloneForm(f => ({ ...f, description: e.target.value }))} />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Review Frequency (days)</Label>
                <Input type="number" value={cloneForm.review_frequency_days} onChange={(e) => setCloneForm(f => ({ ...f, review_frequency_days: e.target.value }))} />
              </div>
              <div>
                <Label>Tags (comma-separated)</Label>
                <Input value={cloneForm.tags} onChange={(e) => setCloneForm(f => ({ ...f, tags: e.target.value }))} />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCloneTemplate(null)}>Cancel</Button>
            <Button onClick={handleClone} disabled={cloning || !cloneForm.identifier}>
              {cloning ? 'Cloning...' : 'Clone Template'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
