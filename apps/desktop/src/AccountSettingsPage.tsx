import { useEffect, useState, type FormEvent } from "react";
import { changeAccountPassword, rejectSession, sendPasswordVerificationNonce, signOut } from "./auth";
import {
  deleteAccountService, listAccountServices, saveAccountService, testAccountService,
  type AccountServiceStatus, type AccountServiceKind,
} from "./accountServices";
import "./accountSettings.css";

type TranslationProvider = "azure" | "baidu";
const aiDefaultURL = "https://api.openai.com/v1";

export default function AccountSettingsPage({ email }: { email: string }) {
  const [services, setServices] = useState<AccountServiceStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [nonce, setNonce] = useState("");
  const [aiURL, setAIURL] = useState(aiDefaultURL);
  const [aiModel, setAIModel] = useState("");
  const [aiKey, setAIKey] = useState("");
  const [translationProvider, setTranslationProvider] = useState<TranslationProvider>("azure");
  const [translationKey, setTranslationKey] = useState("");
  const [azureRegion, setAzureRegion] = useState("");
  const [baiduAppID, setBaiduAppID] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    void listAccountServices(controller.signal).then(rows => {
      if (controller.signal.aborted) return;
      setServices(rows);
      const ai = rows.find(row => row.kind === "ai");
      if (ai) {
        setAIURL(ai.base_url || aiDefaultURL);
        setAIModel(ai.model || "");
      }
      const translation = rows.find(row => row.kind === "translation");
      if (translation?.provider === "azure" || translation?.provider === "baidu") {
        setTranslationProvider(translation.provider);
        setAzureRegion(translation.region || "");
        setBaiduAppID(translation.app_id || "");
      }
    }).catch(e => {
      if (!controller.signal.aborted) setError(message(e));
    }).finally(() => {
      if (!controller.signal.aborted) setLoading(false);
    });
    return () => controller.abort();
  }, []);

  const aiStatus = services.find(row => row.kind === "ai");
  const translationStatus = services.find(row => row.kind === "translation");

  function message(error: unknown) {
    return error instanceof Error ? error.message : "服务请求失败";
  }

  async function run(action: () => Promise<void>, success: string) {
    if (busy) return;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      await action();
      setNotice(success);
    } catch (e) {
      setError(message(e));
    } finally {
      setBusy(false);
    }
  }

  async function updateService(kind: AccountServiceKind, input: Parameters<typeof saveAccountService>[1]) {
    const status = await saveAccountService(kind, input);
    setServices(previous => [...previous.filter(row => row.kind !== kind), status]);
    if (kind === "ai") setAIKey("");
    else setTranslationKey("");
  }

  function saveAI(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!aiKey || !aiModel.trim() || !aiURL.trim()) {
      setError("请填写 AI API 地址、模型和新 Key。");
      return;
    }
    void run(() => updateService("ai", {
      provider: "openai-compatible", base_url: aiURL.trim().replace(/\/+$/, ""),
      model: aiModel.trim(), api_key: aiKey,
    }), "AI 云端凭据已保存。请单独点击测试连接确认模型兼容性。");
  }

  function saveTranslation(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!translationKey) {
      setError("请输入新的翻译服务 Key。");
      return;
    }
    void run(() => updateService("translation", {
      provider: translationProvider, api_key: translationKey,
      ...(translationProvider === "azure" ? { region: azureRegion.trim() } : { app_id: baiduAppID.trim() }),
    }), "中文翻译凭据已保存，未自动发起付费请求。");
  }

  function remove(kind: AccountServiceKind) {
    if (!window.confirm("确定删除当前账号绑定的" + (kind === "ai" ? "AI" : "翻译") + "服务凭据吗？")) return;
    void run(async () => {
      await deleteAccountService(kind);
      setServices(previous => previous.filter(item => item.kind !== kind));
      if (kind === "ai") setAIKey("");
      else setTranslationKey("");
    }, "凭据已从云端删除，此账号后续调用将需要重新绑定。");
  }

  function test(kind: AccountServiceKind) {
    if (!window.confirm("这将真实调用服务商接口，可能产生少量费用。确定开始测试吗？")) return;
    void run(async () => {
      const result = await testAccountService(kind);
      if (!result.ok) throw new Error("提供方未确认成功");
    }, "连接测试通过；此结果不代表商品文案质量已人工验收。");
  }

  function updatePassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!currentPassword || newPassword.length < 8 || newPassword !== confirmation) {
      setError("请输入当前密码、至少 8 位新密码，并确保两次新密码一致。");
      return;
    }
    void run(async () => {
      await changeAccountPassword(currentPassword, newPassword, nonce);
      setCurrentPassword("");
      setNewPassword("");
      setConfirmation("");
      setNonce("");
      window.alert("密码修改成功。为保护账号，请用新密码重新登录。");
      // Local auth-lost notification must not wait for a stalled GoTrue logout.
      void signOut();
      rejectSession();
    }, "密码已修改");
  }

  return <section className="account-settings" aria-label="账号管理">
    <div className="page-heading">
      <div><h1>账号管理</h1><p>修改登录密码，绑定专属于当前账号的云端 AI 与翻译服务。</p></div>
    </div>
    {error && <p className="account-message error" role="alert">{error}</p>}
    {notice && <p className="account-message" role="status">{notice}</p>}

    <article className="account-card">
      <h2>账户安全</h2>
      <p className="account-help">登录账号：<strong>{email}</strong>。密码由 Supabase Auth 管理；当前密码是否强制校验取决于 Supabase Auth 的安全配置，建议管理员开启「Require current password」。</p>
      <form onSubmit={updatePassword} className="account-form">
        <label>当前密码（登录服务启用校验时验证）<input type="password" autoComplete="current-password" value={currentPassword}
          onChange={e => setCurrentPassword(e.target.value)} disabled={busy} required /></label>
        <label>新密码<input type="password" autoComplete="new-password" value={newPassword} minLength={8}
          onChange={e => setNewPassword(e.target.value)} disabled={busy} required /></label>
        <label>确认新密码<input type="password" autoComplete="new-password" value={confirmation}
          onChange={e => setConfirmation(e.target.value)} disabled={busy} required /></label>
        <label>邮箱二次验证码（账号要求时填写）
          <input type="text" inputMode="numeric" autoComplete="one-time-code" value={nonce}
            onChange={e => setNonce(e.target.value)} disabled={busy} maxLength={16} placeholder="选填" /></label>
        <div className="account-actions">
          <button type="button" className="refresh-button" disabled={busy} onClick={() =>
            void run(sendPasswordVerificationNonce, "已请求发送邮箱验证码，请检查邮箱。")}>发送二次验证码</button>
          <button type="submit" className="primary-button" disabled={busy}>修改密码</button>
        </div>
      </form>
    </article>

    <article className="account-card">
      <h2>AI 模型服务</h2>
      <p className="account-help">使用与 OpenAI Chat Completions 兼容的云 API。地址必须由服务器管理员允许，密钥只在云端加密保存。修改模型或 API 地址也需要重新填写完整 Key。</p>
      <p className="account-connection" role="status">{loading ? "正在读取配置…" :
        aiStatus?.needs_rebind ? "密钥需要重新绑定（服务端加密主密钥已更换）" : aiStatus?.configured ? "已绑定 · " + aiStatus.provider + " / " + (aiStatus.model || "模型未提供") : "尚未绑定"}</p>
      <form onSubmit={saveAI} className="account-form">
        <label>API 地址（HTTPS）<input value={aiURL} disabled={busy} onChange={e => setAIURL(e.target.value)}
          placeholder={aiDefaultURL} autoComplete="off" required /></label>
        <label>模型名称<input value={aiModel} disabled={busy} onChange={e => setAIModel(e.target.value)}
          placeholder="填写服务商支持的模型 ID" autoComplete="off" required /></label>
        <label>新 API Key<input type="password" value={aiKey} disabled={busy} onChange={e => setAIKey(e.target.value)}
          placeholder={aiStatus ? "已绑定的 Key 不会显示，填写新 Key 才会替换" : "输入 API Key"}
          autoComplete="new-password" required /></label>
        <div className="account-actions">
          <button type="submit" className="primary-button" disabled={busy || loading}>保存 AI 服务</button>
          <button type="button" className="refresh-button" disabled={busy || !aiStatus}
            onClick={() => test("ai")}>测试连接</button>
          <button type="button" className="refresh-button" disabled={busy || !aiStatus}
            onClick={() => remove("ai")}>删除绑定</button>
        </div>
      </form>
    </article>

    <article className="account-card">
      <h2>中文翻译服务</h2>
      <p className="account-help">用于订单标题、类目和属性的中文阅读，不代替商品详情 EN/MX/BR AI 生成。</p>
      <p className="account-connection" role="status">{loading ? "正在读取配置…" :
        translationStatus?.needs_rebind ? "密钥需要重新绑定（服务端加密主密钥已更换）" : translationStatus?.configured ? "已绑定 · " + (translationStatus.provider === "azure" ? "Azure Translator" : "百度翻译") : "尚未绑定"}</p>
      <form onSubmit={saveTranslation} className="account-form">
        <label>翻译服务商<select value={translationProvider} disabled={busy}
          onChange={e => setTranslationProvider(e.target.value as TranslationProvider)}>
          <option value="azure">Azure Translator</option><option value="baidu">百度翻译</option>
        </select></label>
        {translationProvider === "azure" ? <label>Azure Region<input value={azureRegion} disabled={busy}
          onChange={e => setAzureRegion(e.target.value)} placeholder="例如 eastus" required /></label>
          : <label>百度 APP ID<input value={baiduAppID} disabled={busy}
            onChange={e => setBaiduAppID(e.target.value)} placeholder="百度翻译平台 APP ID" required /></label>}
        <label>{translationProvider === "azure" ? "Azure API Key" : "百度密钥"}
          <input type="password" value={translationKey} disabled={busy}
            onChange={e => setTranslationKey(e.target.value)} autoComplete="new-password"
            placeholder={translationStatus ? "已绑定 Key 不会回传，填入新 Key 才替换" : "请输入服务商密钥"} required /></label>
        <div className="account-actions">
          <button type="submit" className="primary-button" disabled={busy || loading}>保存翻译服务</button>
          <button type="button" className="refresh-button" disabled={busy || !translationStatus}
            onClick={() => test("translation")}>测试翻译</button>
          <button type="button" className="refresh-button" disabled={busy || !translationStatus}
            onClick={() => remove("translation")}>删除绑定</button>
        </div>
      </form>
    </article>
  </section>;
}
