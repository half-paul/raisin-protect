'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import { ArrowLeft, Save, Eye, FileText, AlertTriangle } from 'lucide-react';
import {
  Policy, getPolicy, updatePolicy, createPolicyVersion,
} from '@/lib/api';

export default function PolicyEditorPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { hasRole } = useAuth();
  const canEdit = hasRole('ciso', 'compliance_manager', 'security_engineer');

  const [policy, setPolicy] = useState<Policy | null>(null);
  const [loading, setLoading] = useState(true);

  // Editor state
  const [content, setContent] = useState('');
  const [showPreview, setShowPreview] = useState(false);

  // Metadata edit
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');

  // Save as new version dialog
  const [showSave, setShowSave] = useState(false);
  const [changeSummary, setChangeSummary] = useState('');
  const [changeType, setChangeType] = useState('minor');
  const [contentSummary, setContentSummary] = useState('');
  const [saving, setSaving] = useState(false);

  const fetchPolicy = useCallback(async () => {
    try {
      setLoading(true);
      const res = await getPolicy(id);
      setPolicy(res.data);
      setContent(res.data.current_version?.content || '');
      setTitle(res.data.title);
      setDescription(res.data.description || '');
    } catch (err) {
      console.error('Failed to fetch policy:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { fetchPolicy(); }, [fetchPolicy]);

  const handleSaveVersion = async () => {
    if (!changeSummary.trim()) return;
    try {
      setSaving(true);
      // Update metadata if changed
      if (policy && (title !== policy.title || description !== (policy.description || ''))) {
        await updatePolicy(id, {
          title: title || undefined,
          description: description || undefined,
        });
      }
      // Create new version
      await createPolicyVersion(id, {
        content,
        content_format: 'html',
        content_summary: contentSummary || undefined,
        change_summary: changeSummary,
        change_type: changeType,
      });
      setShowSave(false);
      router.push(`/policies/${id}`);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center py-20 text-muted-foreground">Loading...</div>;
  }
  if (!policy) {
    return <div className="flex items-center justify-center py-20 text-muted-foreground">Policy not found</div>;
  }
  if (!canEdit) {
    return <div className="flex items-center justify-center py-20 text-muted-foreground">You don&apos;t have permission to edit policies.</div>;
  }
  if (policy.status === 'archived') {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <AlertTriangle className="h-8 w-8 text-yellow-500" />
        <p className="text-muted-foreground">Archived policies cannot be edited.</p>
        <Button variant="outline" onClick={() => router.push(`/policies/${id}`)}>Back to Policy</Button>
      </div>
    );
  }

  const wordCount = content.replace(/<[^>]*>/g, ' ').split(/\s+/).filter(Boolean).length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push(`/policies/${id}`)}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="font-mono text-sm text-muted-foreground">{policy.identifier}</span>
              <Badge variant="secondary" className="text-xs">Editing</Badge>
            </div>
            <h1 className="text-xl font-bold">Edit Policy Content</h1>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">{wordCount} words</span>
          <Button variant="outline" onClick={() => setShowPreview(!showPreview)}>
            <Eye className="h-4 w-4 mr-2" /> {showPreview ? 'Editor' : 'Preview'}
          </Button>
          <Button onClick={() => setShowSave(true)}>
            <Save className="h-4 w-4 mr-2" /> Save Version
          </Button>
        </div>
      </div>

      {/* Metadata */}
      <Card>
        <CardContent className="p-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>Title</Label>
              <Input value={title} onChange={(e) => setTitle(e.target.value)} />
            </div>
            <div>
              <Label>Description</Label>
              <Input value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Note about status revert */}
      {(policy.status === 'approved' || policy.status === 'published') && (
        <Card className="border-yellow-500 bg-yellow-50 dark:bg-yellow-950/20">
          <CardContent className="p-4 flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-yellow-600 flex-shrink-0" />
            <p className="text-sm">
              This policy is <strong>{policy.status}</strong>. Saving a new version will revert it to <strong>Draft</strong> status.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Editor / Preview */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <FileText className="h-5 w-5" />
            {showPreview ? 'Preview' : 'HTML Content Editor'}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {showPreview ? (
            <div
              className="prose dark:prose-invert max-w-none min-h-[400px] p-4 border rounded-md"
              dangerouslySetInnerHTML={{ __html: content }}
            />
          ) : (
            <Textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={24}
              className="font-mono text-sm min-h-[400px]"
              placeholder="Enter policy content in HTML format..."
            />
          )}
        </CardContent>
      </Card>

      {/* Save Version Dialog */}
      <Dialog open={showSave} onOpenChange={setShowSave}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Save New Version</DialogTitle>
            <DialogDescription>
              This will create version {(policy.current_version?.version_number || 1) + 1} of this policy.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Change Summary *</Label>
              <Textarea
                placeholder="Describe what changed..."
                value={changeSummary}
                onChange={(e) => setChangeSummary(e.target.value)}
                rows={3}
              />
            </div>
            <div>
              <Label>Change Type</Label>
              <Select value={changeType} onValueChange={setChangeType}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="major">Major — Significant restructuring or new requirements</SelectItem>
                  <SelectItem value="minor">Minor — Section updates, new clauses</SelectItem>
                  <SelectItem value="patch">Patch — Typo fixes, formatting changes</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Content Summary</Label>
              <Input
                placeholder="Brief summary of this version..."
                value={contentSummary}
                onChange={(e) => setContentSummary(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowSave(false)}>Cancel</Button>
            <Button onClick={handleSaveVersion} disabled={saving || !changeSummary.trim()}>
              {saving ? 'Saving...' : 'Save Version'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
