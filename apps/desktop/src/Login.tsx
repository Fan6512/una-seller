import { useState, type FormEvent } from "react";
import { signIn, type AuthSession } from "./auth";

export default function Login({ onSignedIn }: { onSignedIn: (session: AuthSession) => void }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (pending) return;
    setPending(true);
    setError("");
    try {
      onSignedIn(await signIn(email, password));
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败，请重试");
    } finally {
      setPending(false);
    }
  }

  return (
    <main className="login-page">
      <section className="login-card" aria-labelledby="login-title">
        <div className="login-brand"><span className="brand-mark" aria-hidden="true">U</span><span className="brand-name">Una Seller</span></div>
        <div className="login-heading">
          <h1 id="login-title">欢迎回来</h1>
          <p>登录后继续管理你的商品</p>
        </div>
        <form onSubmit={(event) => void submit(event)} className="login-form">
          <label htmlFor="login-email">邮箱地址</label>
          <input id="login-email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="username" placeholder="请输入邮箱地址" maxLength={254} required disabled={pending} autoFocus />
          <label htmlFor="login-password">密码</label>
          <div className="password-field">
            <input id="login-password" type={showPassword ? "text" : "password"} value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password" placeholder="请输入密码" required disabled={pending} />
            <button type="button" aria-label={showPassword ? "隐藏密码" : "显示密码"} onClick={() => setShowPassword((v) => !v)} disabled={pending}>{showPassword ? "隐藏" : "显示"}</button>
          </div>
          {error && <p className="login-error" role="alert">{error}</p>}
          <button type="submit" className="login-submit" disabled={pending || !email.trim() || !password}>{pending ? "正在登录…" : "登录"}</button>
        </form>
        <p className="login-footer">首期仅支持管理员创建的账号</p>
      </section>
    </main>
  );
}
