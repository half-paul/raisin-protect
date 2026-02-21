'use client';

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  Search, BookOpen, FileCheck, Plus,
} from 'lucide-react';
import {
  AuditRequestTemplate, listAuditRequestTemplates, createRequestsFromTemplate, listAudits, Audit,
} from '@/lib/api';
import {
  AUDIT_TYPE_LABELS, TEMPLATE_FRAMEWORK_LABELS,
  REQUEST_PRIORITY_COLORS,
} from '@/components/audit/constants';

export default function PBCTemplatesPage() {
  const { hasRole } = useAuth();
  const canCreate = hasRole('ciso', 'compliance_manager', 'auditor');

  const [templates, setTemplates] = useState<AuditRequestTemplate[]>([]);
  const [audits, setAudits] = useState<Audit[]>([]);
  const [loading, setLoading] = useState(true);

  // Filters
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [frameworkFilter, setFrameworkFilter] = useState('');

  // Selection
  const [selected, setSelected] = useState<Set<string>>(new Set());

  // Bulk create dialog
  const [bulkOpen, setBulkOpen] = useState(false);
  const [bulkAudit, setBulkAudit] = useState('');
  const [bulkDueDate, setBulkDueDate] = useState('');
  const [bulkPrefix, setBulkPrefix] = useState('PBC');
  const [bulkCreating, setBulkCreating] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const p: Record<string, string> = {};
      if (search) p.search = search;
      if (typeFilter) p.audit_type = typeFilter;
      if (frameworkFilter) p.framework = frameworkFilter;

      const [tmplRes, auditsRes] = await Promise.all([
        listAuditRequestTemplates(p),
        listAudits({ per_page: '100', status: 'planning' }).catch(() => ({ data: [] })),
      ]);
      setTemplates(tmplRes.data);
      setAudits(auditsRes.data);
    } catch (err) {
      console.error('Failed to fetch templates:', err);
    } finally {
      setLoading(false);
    }
  }, [search, typeFilter, frameworkFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const toggleSelect = (id: string) => {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id); else next.add(id);
    setSelected(next);
  };

  const toggleAll = () => {
    if (selected.size === templates.length) {
      setSelected(new Set());
    } else {
      setSelected(new Set(templates.map(t => t.id)));
    }
  };

  const handleBulkCreate = async () => {
    if (!bulkAudit || selected.size === 0) return;
    try {
      setBulkCreating(true);
      const result = await createRequestsFromTemplate(bulkAudit, {
        template_ids: Array.from(selected),
        default_due_date: bulkDueDate || undefined,
        auto_number: true,
        number_prefix: bulkPrefix || 'PBC',
      });
      alert(`Created ${result.data.created} evidence requests!`);
      setBulkOpen(false);
      setSelected(new Set());
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create requests');
    } finally {
      setBulkCreating(false);
    }
  };

  // Group by framework
  const grouped = templates.reduce<Record<string, AuditRequestTemplate[]>>((acc, t) => {
    const key = t.framework || 'other';
    if (!acc[key]) acc[key] = [];
    acc[key].push(t);
    return acc;
  }, {});

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">PBC Template Library</h1>
          <p className="text-sm text-muted-foreground">Prepared-By-Client request templates for audit engagements</p>
        </div>
        {canCreate && selected.size > 0 && (
          <Button onClick={() => setBulkOpen(true)}>
            <Plus className="h-4 w-4 mr-2" /> Create {selected.size} Requests
          </Button>
        )}
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-4">
            <div className="flex-1 min-w-[200px]">
              <Label className="text-xs">Search</Label>
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input placeholder="Search templates..." value={searchInput} onChange={(e) => setSearchInput(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); } }} className="pl-8" />
              </div>
            </div>
            <div className="w-[200px]">
              <Label className="text-xs">Audit Type</Label>
              <Select value={typeFilter} onValueChange={(v) => setTypeFilter(v === 'all' ? '' : v)}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All types</SelectItem>
                  {Object.entries(AUDIT_TYPE_LABELS).map(([k, v]) => (<SelectItem key={k} value={k}>{v}</SelectItem>))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[160px]">
              <Label className="text-xs">Framework</Label>
              <Select value={frameworkFilter} onValueChange={(v) => setFrameworkFilter(v === 'all' ? '' : v)}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All frameworks</SelectItem>
                  {Object.entries(TEMPLATE_FRAMEWORK_LABELS).map(([k, v]) => (<SelectItem key={k} value={k}>{v}</SelectItem>))}
                </SelectContent>
              </Select>
            </div>
            <Button variant="outline" size="sm" onClick={toggleAll}>
              {selected.size === templates.length ? 'Deselect All' : 'Select All'}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Templates grouped by framework */}
      {loading ? (
        <div className="text-center py-16 text-muted-foreground">Loading templates...</div>
      ) : templates.length === 0 ? (
        <div className="text-center py-16 text-muted-foreground">No templates found</div>
      ) : (
        Object.entries(grouped).sort(([a], [b]) => a.localeCompare(b)).map(([framework, items]) => (
          <Card key={framework}>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <BookOpen className="h-4 w-4" />
                {TEMPLATE_FRAMEWORK_LABELS[framework] || framework} ({items.length})
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {items.map((tmpl) => (
                <div
                  key={tmpl.id}
                  className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${selected.has(tmpl.id) ? 'bg-primary/5 border-primary/30' : 'hover:bg-muted/50'}`}
                  onClick={() => toggleSelect(tmpl.id)}
                >
                  <Checkbox
                    checked={selected.has(tmpl.id)}
                    onCheckedChange={() => toggleSelect(tmpl.id)}
                    className="mt-1"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-sm">{tmpl.title}</p>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium ${REQUEST_PRIORITY_COLORS[tmpl.default_priority]}`}>
                        {tmpl.default_priority}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground mt-0.5">{tmpl.description}</p>
                    {tmpl.tags.length > 0 && (
                      <div className="flex gap-1 flex-wrap mt-1">
                        {tmpl.tags.map(t => <Badge key={t} variant="secondary" className="text-[10px]">{t}</Badge>)}
                      </div>
                    )}
                  </div>
                  <Badge variant="outline" className="text-[10px] shrink-0">{tmpl.category}</Badge>
                </div>
              ))}
            </CardContent>
          </Card>
        ))
      )}

      {/* Bulk Create Dialog */}
      <Dialog open={bulkOpen} onOpenChange={setBulkOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Requests from Templates</DialogTitle>
            <DialogDescription>{selected.size} templates selected. Choose an audit to add them to.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <Label>Target Audit *</Label>
              <Select value={bulkAudit} onValueChange={setBulkAudit}>
                <SelectTrigger><SelectValue placeholder="Select an audit..." /></SelectTrigger>
                <SelectContent>
                  {audits.map(a => (<SelectItem key={a.id} value={a.id}>{a.title}</SelectItem>))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Default Due Date</Label>
                <Input type="date" value={bulkDueDate} onChange={(e) => setBulkDueDate(e.target.value)} />
              </div>
              <div>
                <Label>Reference Prefix</Label>
                <Input value={bulkPrefix} onChange={(e) => setBulkPrefix(e.target.value)} placeholder="PBC" />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setBulkOpen(false)}>Cancel</Button>
            <Button onClick={handleBulkCreate} disabled={bulkCreating || !bulkAudit}>
              {bulkCreating ? 'Creating...' : `Create ${selected.size} Requests`}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
