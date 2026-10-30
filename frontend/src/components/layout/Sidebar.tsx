import React from 'react';
import {
  LayoutDashboard,
  FolderKanban,
  Award,
  Users,
  Printer,
  Settings,
  FolderGit2,
} from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';

export type NavItem =
  | 'dashboard'
  | 'projects'
  | 'templates'
  | 'recipients'
  | 'history'
  | 'settings';

interface SidebarProps {
  currentTab: NavItem;
  onSelectTab: (tab: NavItem) => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ currentTab, onSelectTab }) => {
  const { activeProject } = useProjectStore();

  const navItems = [
    { id: 'dashboard' as NavItem, label: 'Dashboard', icon: LayoutDashboard },
    { id: 'projects' as NavItem, label: 'Acara', icon: FolderKanban },
    { id: 'templates' as NavItem, label: 'Templates', icon: Award, disabled: !activeProject },
    { id: 'recipients' as NavItem, label: 'Penerima', icon: Users, disabled: !activeProject },
    { id: 'history' as NavItem, label: 'Siap Cetak', icon: Printer },
    { id: 'settings' as NavItem, label: 'Pengaturan', icon: Settings },
  ];

  return (
    <aside className="w-64 bg-slate-900 text-slate-300 flex flex-col border-r border-slate-800 select-none">
      {/* Brand Header */}
      <div className="h-16 flex items-center px-6 gap-3 border-b border-slate-800">
        <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center text-white font-bold shadow-md">
          CG
        </div>
        <div>
          <h1 className="text-sm font-bold text-white tracking-wide leading-tight">
            Certificate
          </h1>
          <p className="text-xs text-slate-400">Generator Janaiz</p>
        </div>
      </div>

      {/* Navigation Links */}
      <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = currentTab === item.id;
          return (
            <button
              key={item.id}
              disabled={item.disabled}
              onClick={() => onSelectTab(item.id)}
              className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-blue-600 text-white shadow-sm'
                  : item.disabled
                  ? 'opacity-40 cursor-not-allowed text-slate-500'
                  : 'text-slate-400 hover:text-white hover:bg-slate-800'
              }`}
            >
              <Icon className="w-4 h-4 flex-shrink-0" />
              <span>{item.label}</span>
              {item.disabled && (
                <span className="ml-auto text-[10px] uppercase tracking-wider bg-slate-800 text-slate-500 px-1.5 py-0.5 rounded">
                  Pilih Acara
                </span>
              )}
            </button>
          );
        })}
      </nav>

      {/* Active Project Card */}
      <div className="p-4 border-t border-slate-800 bg-slate-950/40">
        <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2 flex items-center gap-1.5">
          <FolderGit2 className="w-3.5 h-3.5 text-blue-400" />
          Acara Aktif
        </div>
        {activeProject ? (
          <div className="bg-slate-800/80 p-2.5 rounded-lg border border-slate-700/60">
            <p className="text-xs font-semibold text-white truncate">
              {activeProject.name}
            </p>
            <p className="text-[11px] text-slate-400 truncate mt-0.5">
              {activeProject.description || 'Tidak ada deskripsi'}
            </p>
          </div>
        ) : (
          <div className="text-xs text-slate-500 italic p-2 bg-slate-900/50 rounded border border-dashed border-slate-800">
            Belum ada acara yang dibuka
          </div>
        )}
      </div>
    </aside>
  );
};
