using System.ComponentModel;
using System.Drawing;
using System.Runtime.CompilerServices;
using System.Windows.Forms;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Detail;
using ZYing.UI;

namespace ZYing.Channel;

[ToolboxItem(false)]
public class Meli新版上传 : EContainer
{
	[CompilerGenerated]
	private sealed class <>c__DisplayClass6_0
	{
		public Meli新版上传 <>4__this;

		public ZClient zc;

		public string s;

		public <>c__DisplayClass6_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}

		internal void <Receiving>b__0()
		{
			<>4__this.Receiving(zc, s);
		}
	}

	public const string Key = "SubmitNew";

	public const string Caption = "跨境站刊登";

	private IContainer components;

	private EPanelScroll main;

	private EPanelBase right;

	private ELabel lblTask;

	private ELabelArea info;

	public static Meli新版上传 Instance => Global.Create(typeof(Meli新版上传)) as Meli新版上传;

	public Meli新版上传()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		main.Font = ETheme.微软雅黑Bold;
		main.Icon = Icos.Get(IcoEnums.article);
		Meli新版上传Detail instance = Meli新版上传Detail.Instance;
		instance.Submit = () =>
		{
			info.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408E9962 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44);
		};
		instance.Receiving = Receiving;
		main.Page = instance;
		lblTask.EIcos = IcoEnums.sys;
	}

	public override void OnActive()
	{
		Meli新版上传Detail.Instance.OnActive();
	}

	private void Receiving(ZClient zc, string s)
	{
		<>c__DisplayClass6_0 CS$<>8__locals7 = new <>c__DisplayClass6_0();
		CS$<>8__locals7.<>4__this = this;
		CS$<>8__locals7.zc = zc;
		CS$<>8__locals7.s = s;
		if (InvokeRequired)
		{
			Invoke(() =>
			{
				CS$<>8__locals7.<>4__this.Receiving(CS$<>8__locals7.zc, CS$<>8__locals7.s);
			});
		}
		else
		{
			info.AppendText(CS$<>8__locals7.s + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB709B7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4));
		}
	}

	protected override void Dispose(bool disposing)
	{
		if (disposing && components != null)
		{
			components.Dispose();
		}
		base.Dispose(disposing);
	}

	private void InitializeComponent()
	{
		this.components = new System.ComponentModel.Container();
		this.main = new ZYing.UI.EPanelScroll();
		this.right = new ZYing.UI.EPanelBase();
		this.info = new ZYing.UI.ELabelArea();
		this.lblTask = new ZYing.UI.ELabel();
		this.right.SuspendLayout();
		base.SuspendLayout();
		this.main.CloseBox = false;
		this.main.Dock = System.Windows.Forms.DockStyle.Fill;
		this.main.Location = new System.Drawing.Point(0, 0);
		this.main.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3224EBB8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		this.main.Padding = new System.Windows.Forms.Padding(12, 6, 12, 10);
		this.main.Size = new System.Drawing.Size(570, 657);
		this.main.TabIndex = 0;
		this.main.TabPadding = new System.Windows.Forms.Padding(3, 13, 0, 9);
		this.main.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224C4084 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2);
		this.right.Controls.Add(this.info);
		this.right.Controls.Add(this.lblTask);
		this.right.Dock = System.Windows.Forms.DockStyle.Right;
		this.right.Location = new System.Drawing.Point(570, 0);
		this.right.MaximumSize = new System.Drawing.Size(600, 500000);
		this.right.MinimumSize = new System.Drawing.Size(400, 0);
		this.right.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x12AF4EC5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3);
		this.right.Padding = new System.Windows.Forms.Padding(10, 6, 10, 10);
		this.right.Size = new System.Drawing.Size(400, 657);
		this.right.TabIndex = 10;
		this.right.TabPadding = new System.Windows.Forms.Padding(3, 13, 0, 9);
		this.info.Dock = System.Windows.Forms.DockStyle.Fill;
		this.info.LineSpacing = 1;
		this.info.Location = new System.Drawing.Point(14, 43);
		this.info.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x867A0F1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		this.info.Padding = new System.Windows.Forms.Padding(6, 8, 6, 4);
		this.info.Size = new System.Drawing.Size(376, 32);
		this.info.TabIndex = 11;
		this.info.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468A5956 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6);
		this.lblTask.Border3D = System.Windows.Forms.Border3DSide.Bottom;
		this.lblTask.Dock = System.Windows.Forms.DockStyle.Top;
		this.lblTask.Location = new System.Drawing.Point(14, 6);
		this.lblTask.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E45AFB9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		this.lblTask.Padding = new System.Windows.Forms.Padding(6, 5, 6, 13);
		this.lblTask.Size = new System.Drawing.Size(376, 37);
		this.lblTask.TabIndex = 11;
		this.lblTask.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1636E5BF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8);
		base.Controls.Add(this.main);
		base.Controls.Add(this.right);
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -1873742850 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		base.Size = new System.Drawing.Size(971, 658);
		this.right.ResumeLayout(false);
		this.right.PerformLayout();
		base.ResumeLayout(false);
	}
}
