// Shell-only navigation. Business modules own their pages, data and permissions.
export type DesktopSection = "operations" | "collections" | "published" | "orders" | "afterSales";
export const desktopSections = [
  { id: "operations", label: "运营中心" },
  { id: "collections", label: "工作台" },
  { id: "published", label: "已刊登商品" },
  { id: "orders", label: "订单管理" },
  { id: "afterSales", label: "售后管理" },
] as const;

function NavIcon({ name }: { name: DesktopSection | "settings" | "account" }) {
  return <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor"
    strokeWidth="1.65" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    {name === "operations" && <><rect x="3.5" y="3.5" width="17" height="17" rx="3"/><path d="M7 15v2m5-6v6m5-9v9"/></>}
    {name === "collections" && <><rect x="3.5" y="4" width="17" height="16" rx="2.5"/><path d="M3.5 9h17M10 9v11"/></>}
    {name === "published" && <><path d="m12 3 8.5 4.5v9L12 21l-8.5-4.5v-9L12 3Z"/><path d="m3.5 7.5 8.5 4.8 8.5-4.8M12 12.3V21"/></>}
    {name === "orders" && <><rect x="5" y="3.5" width="14" height="17" rx="2.5"/><path d="M9 8h6M9 12h6M9 16h4"/></>}
    {name === "afterSales" && <><path d="M20.5 11a8.5 8.5 0 0 1-8.5 8.5H4l1.7-3.8A8.5 8.5 0 1 1 20.5 11Z"/><path d="M8 11h8M8 14h5"/></>}
    {name === "account" && <><circle cx="12" cy="8.5" r="3.2"/><path d="M5.5 20v-1.5a6.5 6.5 0 0 1 13 0V20"/></>}
    {name === "settings" && <><circle cx="12" cy="12" r="3"/><path d="m10 3-.8 2.5-2 .9-2.4-1.2-2 3.4 2 1.7v2.4l-2 1.7 2 3.4 2.4-1.2 2 .9L10 20h4l.8-2.5 2-.9 2.4 1.2 2-3.4-2-1.7v-2.4l2-1.7-2-3.4-2.4 1.2-2-.9L14 3Z"/></>}
  </svg>;
}

type Props = {
  email: string;
  section: DesktopSection | null;
  accountSelected: boolean;
  settingsSelected: boolean;
  updateAvailable: boolean;
  onNavigate: (section: DesktopSection) => void;
  onAccount: () => void;
  onSettings: () => void;
};
export default function DesktopSidebar({ email, section, accountSelected, settingsSelected, updateAvailable,
  onNavigate, onAccount, onSettings }: Props) {
  return <aside className="desktop-sidebar" aria-label="应用侧边栏">
    <div className="desktop-sidebar-brand"><span className="brand-mark" aria-hidden="true">U</span><strong>UNA Seller</strong></div>
    <nav className="desktop-sidebar-nav" aria-label="主导航">
      {desktopSections.map(item => <button key={item.id} type="button" className="desktop-nav-item"
        aria-current={section === item.id ? "page" : undefined} aria-label={item.label} title={item.label}
        onClick={() => onNavigate(item.id)}>
        <NavIcon name={item.id}/><span>{item.label}</span>
      </button>)}
    </nav>
    <div className="desktop-sidebar-footer">
      <button type="button" className="desktop-profile" aria-label="个人中心" title={email || "个人中心"}
        aria-current={accountSelected ? "page" : undefined} onClick={onAccount}>
        <span className="desktop-avatar"><NavIcon name="account"/></span>
        <span className="desktop-profile-text"><strong>个人中心</strong><small>{email || "当前用户"}</small></span>
      </button>
      <button type="button" className="desktop-settings-icon" aria-label={updateAvailable ? "设置，有可用更新" : "设置"}
        title={updateAvailable ? "设置 · 有更新" : "设置"} aria-current={settingsSelected ? "page" : undefined}
        onClick={onSettings}><NavIcon name="settings"/>{updateAvailable && <span className="desktop-update-dot" aria-hidden="true"/>}</button>
    </div>
  </aside>;
}
