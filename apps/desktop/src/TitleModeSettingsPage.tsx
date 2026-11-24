import { setSiteTitleMode, useSiteTitleMode } from "./titleModePreference";

export default function TitleModeSettingsPage() {
  const mode = useSiteTitleMode();
  return <section className="account-settings" aria-label="刊登与标题偏好">
    <div className="page-heading">
      <div><h1>刊登与标题偏好</h1><p>选择商品编辑器的标题展示方式，不改变 Mercado Libre 店铺的官方刊登模型。</p></div>
    </div>
    <article className="account-card">
      <h2>站点标题处理方式</h2>
      <fieldset className="account-form">
        <legend>选择编辑样式</legend>
        <label>
          <input type="radio" name="site-title-mode" value="platform_auto"
            checked={mode === "platform_auto"} onChange={() => setSiteTitleMode("platform_auto")} />
          UP 平台自动生成站点标题（仅填写英文主标题）
        </label>
        <p className="account-help">商品编辑器只显示英文 Family Name / 主标题输入；西语、葡语商品描述仍可编辑。</p>
        <label>
          <input type="radio" name="site-title-mode" value="localized"
            checked={mode === "localized"} onChange={() => setSiteTitleMode("localized")} />
          自定义站点标题（英语 EN / 墨西哥西语 MX / 巴西葡语 BR）
        </label>
        <p className="account-help">商品编辑器显示三语标题切换、字符统计与复制功能，保留每个站点的 SEO 文案。</p>
      </fieldset>
      <p className="account-help" role="status">
        当前选择：{mode === "platform_auto" ? "平台自动生成（英文标题模式）" : "自定义站点标题（三语模式）"}。
        设置仅对当前客户端的编辑界面生效，切换不会覆盖或删除任何已保存的标题、描述及变体。
      </p>
      <p className="account-help">
        注意：CBT C1 发布接口尚未接通。UP 账号具体能否提交自定义站点标题，需要按官方当前接口及账号能力验证；
        本设置不是 Legacy / UP 刊登模式开关，也不会自动执行创建或改标题操作。
      </p>
    </article>
  </section>;
}
