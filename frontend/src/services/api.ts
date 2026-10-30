import * as AppBindings from '../../bindings/certificate-generator/app/app';
import type {
  Project,
  Template,
  Recipient,
  GenerationJob,
  GenerationJobItem,
  BatchConfig,
  AppSettings,
  StorageInfo,
  CreateProjectInput,
  DuplicateProjectInput,
  CreateTemplateInput,
  ColumnMapping,
  ImportSummary,
  RecipientFilter,
  PaginatedRecipients,
  ConflictStrategy,
  BackupManifest,
  RestoreOptions,
  RestoreResult,
} from '../../bindings/certificate-generator/internal/domain/models';
import type {
  FileMetadata,
  TablePreview,
} from '../../bindings/certificate-generator/internal/importer/models';

export interface IApiClient {
  createProject(input: CreateProjectInput): Promise<Project | null>;
  getProject(id: string): Promise<Project | null>;
  listProjects(includeArchived: boolean): Promise<Project[]>;
  recentProjects(limit: number): Promise<Project[]>;
  renameProject(id: string, newName: string, description: string): Promise<Project | null>;
  duplicateProject(input: DuplicateProjectInput): Promise<Project | null>;
  archiveProject(id: string, archive: boolean): Promise<void>;
  deleteProject(id: string): Promise<void>;
  getStorageInfo(): Promise<StorageInfo | null>;
  getSettings(): Promise<AppSettings | null>;
  updateSettings(settings: AppSettings): Promise<void>;

  // Template methods
  createTemplate(input: CreateTemplateInput): Promise<Template | null>;
  getTemplate(id: string): Promise<Template | null>;
  listTemplates(projectId: string): Promise<Template[]>;
  updateTemplate(id: string, name: string, docJson: string): Promise<Template | null>;
  deleteTemplate(id: string): Promise<void>;
  setDefaultTemplate(projectId: string, templateId: string): Promise<void>;
  renderPreviewPdf(docJson: string, variables: Record<string, string>): Promise<string>;

  // Recipient methods
  inspectImportFile(filePath: string): Promise<FileMetadata | null>;
  previewImportFile(filePath: string, sheetName: string, maxRows: number): Promise<TablePreview | null>;
  validateImportMapping(filePath: string, sheetName: string, mapping: ColumnMapping): Promise<ImportSummary | null>;
  importRecipients(
    projectId: string,
    filePath: string,
    sheetName: string,
    mapping: ColumnMapping,
    strategy: ConflictStrategy
  ): Promise<ImportSummary | null>;
  createRecipient(rec: Recipient): Promise<Recipient | null>;
  getRecipient(id: string): Promise<Recipient | null>;
  updateRecipient(rec: Recipient): Promise<Recipient | null>;
  deleteRecipient(id: string): Promise<void>;
  deleteRecipients(ids: string[]): Promise<void>;
  deleteAllRecipients(projectId: string): Promise<void>;
  listRecipients(filter: RecipientFilter): Promise<PaginatedRecipients | null>;
  countRecipients(projectId: string): Promise<number>;

  // Batch Job methods
  startBatchGeneration(cfg: BatchConfig): Promise<GenerationJob | null>;
  cancelBatchJob(jobId: string): Promise<void>;
  retryFailedBatchItems(jobId: string): Promise<GenerationJob | null>;
  getBatchJob(id: string): Promise<GenerationJob | null>;
  listBatchJobs(projectId: string): Promise<GenerationJob[]>;
  listBatchJobItems(jobId: string): Promise<GenerationJobItem[]>;
  deleteBatchJob(id: string): Promise<void>;

  // Backup & Restore methods
  backupProject(projectId: string, outputPath: string): Promise<BackupManifest | null>;
  backupAll(outputPath: string): Promise<BackupManifest | null>;
  restoreProject(opts: RestoreOptions): Promise<RestoreResult | null>;
  openFile(filePath: string): Promise<void>;
  openDirectory(path: string): Promise<void>;
  selectOutputDirectory(): Promise<string>;
  getDefaultExportDirectory(projectName: string): Promise<string>;
}

// In-memory mock client used in testing or outside Wails desktop runtime
export class MockApiClient implements IApiClient {
  private projects: Project[] = [];
  private templates: Template[] = [];
  private recipients: Recipient[] = [];
  private jobs: GenerationJob[] = [];
  private jobItems: GenerationJobItem[] = [];
  private settings: AppSettings = {
    defaultOutputDir: '',
    theme: 'system',
    language: 'id',
    autoSaveIntervalSec: 10,
  } as AppSettings;

  constructor(
    initialProjects: Project[] = [],
    initialTemplates: Template[] = [],
    initialRecipients: Recipient[] = [],
    initialJobs: GenerationJob[] = []
  ) {
    this.projects = [...initialProjects];
    this.templates = [...initialTemplates];
    this.recipients = [...initialRecipients];
    this.jobs = [...initialJobs];
  }

  async createProject(input: CreateProjectInput): Promise<Project | null> {
    const newProj = {
      id: 'mock-' + Math.random().toString(36).substring(2, 9),
      name: input.name,
      description: input.description || '',
      eventDate: input.eventDate || '',
      eventLocation: input.eventLocation || '',
      eventAddress: input.eventAddress || '',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      isArchived: false,
    } as unknown as Project;
    this.projects.unshift(newProj);
    return newProj;
  }

  async getProject(id: string): Promise<Project | null> {
    return this.projects.find((p) => p.id === id) || null;
  }

  async listProjects(includeArchived: boolean): Promise<Project[]> {
    if (includeArchived) {
      return [...this.projects];
    }
    return this.projects.filter((p) => !p.isArchived);
  }

  async recentProjects(limit: number): Promise<Project[]> {
    return this.projects.filter((p) => !p.isArchived).slice(0, limit);
  }

  async renameProject(id: string, newName: string, description: string): Promise<Project | null> {
    const proj = this.projects.find((p) => p.id === id);
    if (!proj) return null;
    proj.name = newName;
    proj.description = description;
    proj.updatedAt = new Date().toISOString() as any;
    return proj;
  }

  async duplicateProject(input: DuplicateProjectInput): Promise<Project | null> {
    const source = this.projects.find((p) => p.id === input.sourceProjectId);
    if (!source) throw new Error('Source project not found');

    const dup = {
      id: 'mock-' + Math.random().toString(36).substring(2, 9),
      name: input.newName,
      description: source.description,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      isArchived: false,
    } as unknown as Project;
    this.projects.unshift(dup);
    return dup;
  }

  async archiveProject(id: string, archive: boolean): Promise<void> {
    const proj = this.projects.find((p) => p.id === id);
    if (proj) {
      proj.isArchived = archive;
    }
  }

  async deleteProject(id: string): Promise<void> {
    this.projects = this.projects.filter((p) => p.id !== id);
    this.templates = this.templates.filter((t) => t.projectId !== id);
    this.recipients = this.recipients.filter((r) => r.projectId !== id);
    this.jobs = this.jobs.filter((j) => j.projectId !== id);
  }

  async getStorageInfo(): Promise<StorageInfo | null> {
    const total = this.projects.length;
    const active = this.projects.filter((p) => !p.isArchived).length;
    return {
      dataDir: '/Users/mock/CertificateGenerator',
      databaseSizeByte: 40960,
      projectsSizeByte: 1048576,
      totalProjects: total,
      activeProjects: active,
      archivedProjects: total - active,
    } as unknown as StorageInfo;
  }

  async getSettings(): Promise<AppSettings | null> {
    return { ...this.settings };
  }

  async updateSettings(settings: AppSettings): Promise<void> {
    this.settings = { ...settings };
  }

  async createTemplate(input: CreateTemplateInput): Promise<Template | null> {
    const newTmpl = {
      id: 'tmpl-' + Math.random().toString(36).substring(2, 9),
      projectId: input.projectId,
      name: input.name,
      documentJson:
        input.documentJson ||
        JSON.stringify({
          schemaVersion: 1,
          page: {
            widthMm: 297,
            heightMm: 210,
            orientation: 'landscape',
            backgroundColor: '#FFFFFF',
          },
          elements: [],
        }),
      schemaVersion: 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    } as unknown as Template;
    this.templates.unshift(newTmpl);
    return newTmpl;
  }

  async getTemplate(id: string): Promise<Template | null> {
    return this.templates.find((t) => t.id === id) || null;
  }

  async listTemplates(projectId: string): Promise<Template[]> {
    return this.templates.filter((t) => t.projectId === projectId);
  }

  async updateTemplate(id: string, name: string, docJson: string): Promise<Template | null> {
    const tmpl = this.templates.find((t) => t.id === id);
    if (!tmpl) return null;
    tmpl.name = name;
    tmpl.documentJson = docJson;
    tmpl.updatedAt = new Date().toISOString() as any;
    return tmpl;
  }

  async deleteTemplate(id: string): Promise<void> {
    this.templates = this.templates.filter((t) => t.id !== id);
  }

  async setDefaultTemplate(projectId: string, templateId: string): Promise<void> {
    for (const t of this.templates) {
      if (t.projectId === projectId) {
        (t as any).isDefault = t.id === templateId;
      }
    }
  }

  async renderPreviewPdf(_docJson: string, _variables: Record<string, string>): Promise<string> {
    return 'JVBERi0xLjQK...';
  }

  async inspectImportFile(_filePath: string): Promise<FileMetadata | null> {
    return {
      fileName: 'recipients_sample.xlsx',
      fileSize: 10240,
      fileType: 'xlsx',
      worksheets: ['Sheet1', 'Peserta Lanjutan'],
    } as unknown as FileMetadata;
  }

  async previewImportFile(_filePath: string, _sheetName: string, maxRows: number): Promise<TablePreview | null> {
    return {
      headers: ['Nama Lengkap', 'Email', 'Organisasi', 'Acara', 'Tanggal', 'No Sertifikat', 'Nilai'],
      rows: [
        ['Budi Santoso, S.Kom.', 'budi@mail.com', 'PT Maju Jaya', 'Seminar AI', '30 Oktober 2026', 'CERT-001', '90'],
        ['Siti Aminah, M.T.', 'siti@kampus.id', 'Univ Indonesia', 'Seminar AI', '30 Oktober 2026', 'CERT-002', '95'],
      ].slice(0, maxRows),
      totalRows: 2,
    } as unknown as TablePreview;
  }

  async validateImportMapping(_filePath: string, _sheetName: string, _mapping: ColumnMapping): Promise<ImportSummary | null> {
    return {
      totalRows: 2,
      validRows: 2,
      invalidRows: 0,
      importedRows: 0,
      skippedRows: 0,
      updatedRows: 0,
      errors: [],
    } as unknown as ImportSummary;
  }

  async importRecipients(
    projectId: string,
    _filePath: string,
    _sheetName: string,
    _mapping: ColumnMapping,
    _strategy: ConflictStrategy
  ): Promise<ImportSummary | null> {
    const mockRec1 = {
      id: 'rec-' + Math.random().toString(36).substring(2, 9),
      projectId,
      fullName: 'Budi Santoso, S.Kom.',
      email: 'budi@mail.com',
      organization: 'PT Maju Jaya',
      eventName: 'Seminar AI',
      issueDate: '2026-10-30',
      certificateId: 'CERT-001',
      customFieldsJson: '{"score":"90"}',
      createdAt: new Date().toISOString() as any,
      updatedAt: new Date().toISOString() as any,
    } as unknown as Recipient;

    this.recipients.push(mockRec1);

    return {
      totalRows: 1,
      validRows: 1,
      invalidRows: 0,
      importedRows: 1,
      skippedRows: 0,
      updatedRows: 0,
      errors: [],
    } as unknown as ImportSummary;
  }

  async createRecipient(rec: Recipient): Promise<Recipient | null> {
    const newRec = {
      ...rec,
      id: 'rec-' + Math.random().toString(36).substring(2, 9),
      createdAt: new Date().toISOString() as any,
      updatedAt: new Date().toISOString() as any,
    };
    this.recipients.push(newRec);
    return newRec;
  }

  async getRecipient(id: string): Promise<Recipient | null> {
    return this.recipients.find((r) => r.id === id) || null;
  }

  async updateRecipient(rec: Recipient): Promise<Recipient | null> {
    const index = this.recipients.findIndex((r) => r.id === rec.id);
    if (index === -1) return null;
    this.recipients[index] = { ...rec, updatedAt: new Date().toISOString() as any };
    return this.recipients[index];
  }

  async deleteRecipient(id: string): Promise<void> {
    this.recipients = this.recipients.filter((r) => r.id !== id);
  }

  async deleteRecipients(ids: string[]): Promise<void> {
    const idSet = new Set(ids);
    this.recipients = this.recipients.filter((r) => !idSet.has(r.id));
  }

  async deleteAllRecipients(projectId: string): Promise<void> {
    this.recipients = this.recipients.filter((r) => r.projectId !== projectId);
  }

  async listRecipients(filter: RecipientFilter): Promise<PaginatedRecipients | null> {
    let items = this.recipients.filter((r) => r.projectId === filter.projectId);
    if (filter.search?.trim()) {
      const q = filter.search.toLowerCase();
      items = items.filter(
        (r) =>
          r.fullName.toLowerCase().includes(q) ||
          (r.email && r.email.toLowerCase().includes(q)) ||
          (r.certificateId && r.certificateId.toLowerCase().includes(q))
      );
    }
    const page = filter.page || 1;
    const pageSize = filter.pageSize || 25;
    const total = items.length;
    const offset = (page - 1) * pageSize;
    const pageItems = items.slice(offset, offset + pageSize);
    const totalPages = Math.max(1, Math.ceil(total / pageSize));

    return {
      items: pageItems,
      total,
      page,
      pageSize,
      totalPages,
    } as unknown as PaginatedRecipients;
  }

  async countRecipients(projectId: string): Promise<number> {
    return this.recipients.filter((r) => r.projectId === projectId).length;
  }

  // Batch Job methods
  async startBatchGeneration(cfg: BatchConfig): Promise<GenerationJob | null> {
    const count = cfg.recipientIds && cfg.recipientIds.length > 0 ? cfg.recipientIds.length : this.recipients.length;
    const newJob = {
      id: 'job-' + Math.random().toString(36).substring(2, 9),
      projectId: cfg.projectId,
      templateId: cfg.templateId,
      status: 'completed',
      totalCount: count,
      completedCount: count,
      failedCount: 0,
      outputPath: '/exports/certificates_sample.zip',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    } as unknown as GenerationJob;
    this.jobs.unshift(newJob);
    return newJob;
  }

  async cancelBatchJob(jobId: string): Promise<void> {
    const job = this.jobs.find((j) => j.id === jobId);
    if (job) job.status = 'cancelled';
  }

  async retryFailedBatchItems(jobId: string): Promise<GenerationJob | null> {
    const job = this.jobs.find((j) => j.id === jobId);
    return job || null;
  }

  async getBatchJob(id: string): Promise<GenerationJob | null> {
    return this.jobs.find((j) => j.id === id) || null;
  }

  async listBatchJobs(projectId: string): Promise<GenerationJob[]> {
    return this.jobs.filter((j) => j.projectId === projectId);
  }

  async listBatchJobItems(jobId: string): Promise<GenerationJobItem[]> {
    return this.jobItems.filter((i) => i.jobId === jobId);
  }

  async deleteBatchJob(id: string): Promise<void> {
    this.jobs = this.jobs.filter((j) => j.id !== id);
  }

  // Backup & Restore methods
  async backupProject(projectId: string, _outputPath: string): Promise<BackupManifest | null> {
    return {
      schemaVersion: 1,
      appVersion: '1.0.0',
      backupType: 'project',
      projectId,
      projectName: 'Sample Project',
      createdAt: new Date().toISOString(),
      fileCount: 4,
      totalBytes: 40960,
    } as unknown as BackupManifest;
  }

  async backupAll(_outputPath: string): Promise<BackupManifest | null> {
    return {
      schemaVersion: 1,
      appVersion: '1.0.0',
      backupType: 'full',
      createdAt: new Date().toISOString(),
      fileCount: 12,
      totalBytes: 1048576,
    } as unknown as BackupManifest;
  }

  async restoreProject(opts: RestoreOptions): Promise<RestoreResult | null> {
    const restoredProj = {
      id: 'restored-' + Math.random().toString(36).substring(2, 9),
      name: 'Restored Project',
      description: 'Restored from ' + opts.archiveFilePath,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      isArchived: false,
    } as unknown as Project;
    this.projects.unshift(restoredProj);

    return {
      projectId: restoredProj.id,
      projectName: restoredProj.name,
      status: 'restored',
      message: 'Proyek berhasil dipulihkan',
    } as unknown as RestoreResult;
  }

  async openFile(_filePath: string): Promise<void> {}
  async openDirectory(_path: string): Promise<void> {}
  async selectOutputDirectory(): Promise<string> {
    return '/mock/Downloads/Sertifikat';
  }
  async getDefaultExportDirectory(projectName: string): Promise<string> {
    return `/mock/Downloads/Sertifikat/${projectName || 'Proyek'}`;
  }
}

// Production client forwarding calls to generated Wails bindings
export class WailsApiClient implements IApiClient {
  createProject(input: CreateProjectInput): Promise<Project | null> {
    return AppBindings.CreateProject(input);
  }

  getProject(id: string): Promise<Project | null> {
    return AppBindings.GetProject(id);
  }

  listProjects(includeArchived: boolean): Promise<Project[]> {
    return AppBindings.ListProjects(includeArchived);
  }

  recentProjects(limit: number): Promise<Project[]> {
    return AppBindings.RecentProjects(limit);
  }

  renameProject(id: string, newName: string, description: string): Promise<Project | null> {
    return AppBindings.RenameProject(id, newName, description);
  }

  duplicateProject(input: DuplicateProjectInput): Promise<Project | null> {
    return AppBindings.DuplicateProject(input);
  }

  archiveProject(id: string, archive: boolean): Promise<void> {
    return AppBindings.ArchiveProject(id, archive);
  }

  deleteProject(id: string): Promise<void> {
    return AppBindings.DeleteProject(id);
  }

  getStorageInfo(): Promise<StorageInfo | null> {
    return AppBindings.GetStorageInfo();
  }

  getSettings(): Promise<AppSettings | null> {
    return AppBindings.GetSettings();
  }

  updateSettings(settings: AppSettings): Promise<void> {
    return AppBindings.UpdateSettings(settings);
  }

  createTemplate(input: CreateTemplateInput): Promise<Template | null> {
    return AppBindings.CreateTemplate(input);
  }

  getTemplate(id: string): Promise<Template | null> {
    return AppBindings.GetTemplate(id);
  }

  listTemplates(projectId: string): Promise<Template[]> {
    return AppBindings.ListTemplates(projectId);
  }

  updateTemplate(id: string, name: string, docJson: string): Promise<Template | null> {
    return AppBindings.UpdateTemplate(id, name, docJson);
  }

  deleteTemplate(id: string): Promise<void> {
    return AppBindings.DeleteTemplate(id);
  }

  setDefaultTemplate(projectId: string, templateId: string): Promise<void> {
    return AppBindings.SetDefaultTemplate(projectId, templateId);
  }

  renderPreviewPdf(docJson: string, variables: Record<string, string>): Promise<string> {
    return AppBindings.RenderPreviewPDF(docJson, variables);
  }

  inspectImportFile(filePath: string): Promise<FileMetadata | null> {
    return AppBindings.InspectImportFile(filePath);
  }

  previewImportFile(filePath: string, sheetName: string, maxRows: number): Promise<TablePreview | null> {
    return AppBindings.PreviewImportFile(filePath, sheetName, maxRows);
  }

  validateImportMapping(filePath: string, sheetName: string, mapping: ColumnMapping): Promise<ImportSummary | null> {
    return AppBindings.ValidateImportMapping(filePath, sheetName, mapping);
  }

  importRecipients(
    projectId: string,
    filePath: string,
    sheetName: string,
    mapping: ColumnMapping,
    strategy: ConflictStrategy
  ): Promise<ImportSummary | null> {
    return AppBindings.ImportRecipients(projectId, filePath, sheetName, mapping, strategy);
  }

  createRecipient(rec: Recipient): Promise<Recipient | null> {
    return AppBindings.CreateRecipient(rec);
  }

  getRecipient(id: string): Promise<Recipient | null> {
    return AppBindings.GetRecipient(id);
  }

  updateRecipient(rec: Recipient): Promise<Recipient | null> {
    return AppBindings.UpdateRecipient(rec);
  }

  deleteRecipient(id: string): Promise<void> {
    return AppBindings.DeleteRecipient(id);
  }

  deleteRecipients(ids: string[]): Promise<void> {
    return AppBindings.DeleteRecipients(ids);
  }

  deleteAllRecipients(projectId: string): Promise<void> {
    return AppBindings.DeleteAllRecipients(projectId);
  }

  listRecipients(filter: RecipientFilter): Promise<PaginatedRecipients | null> {
    return AppBindings.ListRecipients(filter);
  }

  countRecipients(projectId: string): Promise<number> {
    return AppBindings.CountRecipients(projectId);
  }

  startBatchGeneration(cfg: BatchConfig): Promise<GenerationJob | null> {
    return AppBindings.StartBatchGeneration(cfg);
  }

  cancelBatchJob(jobId: string): Promise<void> {
    return AppBindings.CancelBatchJob(jobId);
  }

  retryFailedBatchItems(jobId: string): Promise<GenerationJob | null> {
    return AppBindings.RetryFailedBatchItems(jobId);
  }

  getBatchJob(id: string): Promise<GenerationJob | null> {
    return AppBindings.GetBatchJob(id);
  }

  listBatchJobs(projectId: string): Promise<GenerationJob[]> {
    return AppBindings.ListBatchJobs(projectId);
  }

  listBatchJobItems(jobId: string): Promise<GenerationJobItem[]> {
    return AppBindings.ListBatchJobItems(jobId);
  }

  deleteBatchJob(id: string): Promise<void> {
    return AppBindings.DeleteBatchJob(id);
  }

  backupProject(projectId: string, outputPath: string): Promise<BackupManifest | null> {
    return AppBindings.BackupProject(projectId, outputPath);
  }

  backupAll(outputPath: string): Promise<BackupManifest | null> {
    return AppBindings.BackupAll(outputPath);
  }

  restoreProject(opts: RestoreOptions): Promise<RestoreResult | null> {
    return AppBindings.RestoreProject(opts);
  }

  openFile(filePath: string): Promise<void> {
    return AppBindings.OpenFile(filePath);
  }

  openDirectory(path: string): Promise<void> {
    return AppBindings.OpenDirectory(path);
  }

  selectOutputDirectory(): Promise<string> {
    return AppBindings.SelectOutputDirectory();
  }

  getDefaultExportDirectory(projectName: string): Promise<string> {
    return AppBindings.GetDefaultExportDirectory(projectName);
  }
}

// Auto-detect environment
export const isWailsEnvironment =
  typeof window !== 'undefined' &&
  (Boolean((window as any).chrome?.webview?.postMessage) ||
    Boolean((window as any).webkit?.messageHandlers?.['external']?.postMessage) ||
    Boolean((window as any).wails?.invoke));

export const apiClient: IApiClient = isWailsEnvironment
  ? new WailsApiClient()
  : new MockApiClient();
