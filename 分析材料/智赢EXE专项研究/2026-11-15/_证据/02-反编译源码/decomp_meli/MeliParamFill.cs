using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Drawing;
using System.Linq;
using System.Windows.Forms;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.DWin;
using ZYing.Data;
using ZYing.Static;
using ZYing.UI;

namespace ZYing.Channel;

[ToolboxItem(false)]
public class Meli参数补全 : CTableCols
{
	public const string Key = "MeliExtra";

	public const string Caption = "参数补全";

	private IContainer components;

	private ELoader lblTotal;

	private WMeliExtra right;

	private ELabel Subject;

	private ESearch PWord;

	private EDropLinkage PKind;

	private EButton btnNew;

	private EDropLinkage selMeliCate;

	private ESelect PType;

	private EButton 下载分类属性;

	private EButton 导入产品属性;

	public static Meli参数补全 Instance => Global.Create(typeof(Meli参数补全)) as Meli参数补全;

	public Meli参数补全()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		Dock = DockStyle.Fill;
		tbl.Columns.AddRange(new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F21C6E8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0xE865D4E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947), 20, 0, hidden: true), new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68D2C21E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A38382 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010), 200, int.MaxValue), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EADFE2A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x582052DB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd), 70, 0, ETCellStyle.Number), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16352A57 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D35C526 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b), 55, 55, ETCellStyle.OP));
		tbl.RowCreating = Tbl_RowCreating;
		tbl.RowSetting = Tbl_RowSetting;
		tbl.RowChanged = Tbl_RowChanged;
		tbl.CellOPClick = Tbl_OPClick;
		PType.SelectedChanged += DoSelect;
		PType.Items = new EItemCollection(new EItem(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75C8FCC7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206)), new EItem(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A5063E8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39)));
		selMeliCate.Kinder = KindMeli.Instance;
		selMeliCate.SelectedChanged += DoSelect;
		PKind.Kinder = KindMeli.Instance;
		PWord.OK += Filter;
		PWord.Void += DoVoid;
		btnNew.Click += BtnNewClick;
		right.Updated = DoSelect;
		下载分类属性.Click += 下载分类属性_Click;
		导入产品属性.Click += 导入产品属性_Click;
	}

	private void 导入产品属性_Click(object sender, EventArgs e)
	{
		Meli产品属性导入.Instance.Show(this);
	}

	private void 下载分类属性_Click(object sender, EventArgs e)
	{
		if (PKind.Value.IsNullOrEmpty())
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E25E16 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce));
		}
		string fullName = KindMeli.Instance.GetFullName(PKind.Value);
		var source = EPropHelper.GetMeliProp(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6EB048 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535)).GetProps(Valid.ToInt(PKind.Value)).Props.Select((EProp x) => new { x.Key, x.Zh }).ToList();
		List<string> list = new List<string>
		{
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527D0811 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B4B710 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D358672 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b)
		};
		List<string> collection = source.Select(x => x.Zh + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11656813 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca) + x.Key).ToList();
		list.AddRange(collection);
		Global.下载模板(PKind.Value, 150, list.ToArray(), new string[3] { "", "", fullName }, new string[0]);
	}

	public override void Load()
	{
		DoSelect(this, EventArgs.Empty);
	}

	private void Filter(object sender, EventArgs e)
	{
		string word = PWord.Text.Trim();
		tbl.Reset();
		tbl.Like(word, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C25032 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03));
	}

	protected void Tbl_RowCreating(ETRow tr)
	{
		tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DCDF112 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5)].Color = Color.DarkRed;
		((ETCellOP)tr.LastCell).AddRange(new UIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C0C514 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03)));
	}

	private void Tbl_RowSetting(ETRow tr, Record re)
	{
		tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x50251DCF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841)] = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F42F260 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835));
	}

	private async void Tbl_RowChanged(ETRow tr, ETRowEventArgs e)
	{
		await right.LoadAsync(tr.Key);
	}

	private async void Tbl_OPClick(ETCellOP c, UIcon ico)
	{
		if (ico.Name == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB096BB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018))
		{
			if (EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xAC8CC2E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057)))
			{
				ZClient zc = ZClient.Create();
				zc.UI = this;
				zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C964719 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
				await zc.PostAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D646A0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), c.Row.Key);
				if (zc.IsErr)
				{
					EDialog.Show(zc.Message, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408BACC2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44));
				}
				else
				{
					DoSelect(this, EventArgs.Empty);
				}
			}
		}
		else
		{
			tbl.CurrentRow = c.Row;
		}
	}

	private Para DoPara()
	{
		return new Para
		{
			localid = PKind.Value.ToInt(),
			local = selMeliCate.Value.ToInt(),
			type = PType.Value.ToInt()
		};
	}

	private void DoVoid(object sender, EventArgs e)
	{
		head.Reset();
		DoSelect(sender, e);
	}

	private void DoSelect(object sender, EventArgs e)
	{
		DoCommand(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F3305F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1), DoPara());
	}

	private async void DoCommand(string cmd, Para p)
	{
		PWord.Reset();
		ZClient zc = ZClient.Create();
		zc.UI = this;
		zc.Loader = lblTotal;
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA051DBB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd) + cmd;
		await zc.PostAsync(p);
		Record record = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x8660F8B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788)];
		tbl.DataSource = record;
		lblTotal.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x60DDCC65 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97) + record.MaxCount + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x8F8C52E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6);
	}

	private async void BtnNewClick(object sender, EventArgs e)
	{
		if (PKind.Value.IsNullOrEmpty())
		{
			EDialog.Show(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D3FAFC6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F1D52F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1));
			return;
		}
		ZClient zClient = ZClient.Create();
		zClient.UI = this;
		zClient.Loader = lblTotal;
		zClient.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E4C4915 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		await zClient.PostAsync(DoPara());
		DoSelect(sender, e);
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
		this.lblTotal = new ZYing.UI.ELoader();
		this.right = new ZYing.DWin.WMeliExtra();
		this.Subject = new ZYing.UI.ELabel();
		this.PWord = new ZYing.UI.ESearch();
		this.PKind = new ZYing.UI.EDropLinkage();
		this.btnNew = new ZYing.UI.EButton();
		this.selMeliCate = new ZYing.UI.EDropLinkage();
		this.PType = new ZYing.UI.ESelect();
		this.下载分类属性 = new ZYing.UI.EButton();
		this.导入产品属性 = new ZYing.UI.EButton();
		base.main.SuspendLayout();
		base.tail.SuspendLayout();
		base.head.SuspendLayout();
		base.SuspendLayout();
		base.main.Size = new System.Drawing.Size(524, 672);
		base.tail.Controls.Add(this.lblTotal);
		base.tail.Controls.Add(this.PKind);
		base.tail.Controls.Add(this.下载分类属性);
		base.tail.Controls.Add(this.导入产品属性);
		base.tail.Controls.Add(this.btnNew);
		base.tail.Location = new System.Drawing.Point(7, 597);
		base.tail.Size = new System.Drawing.Size(509, 65);
		base.tbl.Location = new System.Drawing.Point(10, 76);
		base.tbl.MaxHeight = 521;
		base.tbl.Size = new System.Drawing.Size(503, 518);
		base.tbl.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x764368D1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
		base.head.Controls.Add(this.Subject);
		base.head.Controls.Add(this.selMeliCate);
		base.head.Controls.Add(this.PType);
		base.head.Controls.Add(this.PWord);
		base.head.Size = new System.Drawing.Size(509, 67);
		this.lblTotal.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.lblTotal.ForeColor = System.Drawing.Color.Black;
		this.lblTotal.Location = new System.Drawing.Point(343, 3);
		this.lblTotal.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9C6F33 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
		this.lblTotal.Padding = new System.Windows.Forms.Padding(5, 3, 2, 3);
		this.lblTotal.Size = new System.Drawing.Size(162, 26);
		this.lblTotal.TabIndex = 10;
		this.lblTotal.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D3DAD8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda);
		this.lblTotal.TextAlign = System.Windows.Forms.HorizontalAlignment.Right;
		this.right.Location = new System.Drawing.Point(534, 0);
		this.right.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10352178 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
		this.right.Padding = new System.Windows.Forms.Padding(6);
		this.right.Size = new System.Drawing.Size(524, 119);
		this.right.TabIndex = 8;
		this.right.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7083DF80 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834);
		this.right.TextColor = System.Drawing.Color.Empty;
		this.Subject.EIcos = ZYing.IcoEnums.article;
		this.Subject.FlowBreak = false;
		this.Subject.Fonts = ZYing.UI.EFonts.微软雅黑Bold;
		this.Subject.IconY = -1;
		this.Subject.LabelColor = System.Drawing.Color.Empty;
		this.Subject.Location = new System.Drawing.Point(3, 3);
		this.Subject.Margin = new System.Windows.Forms.Padding(3, 3, 0, 3);
		this.Subject.MaxHeight = -7;
		this.Subject.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5110C6B7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
		this.Subject.Padding = new System.Windows.Forms.Padding(0, 5, 0, 3);
		this.Subject.Size = new System.Drawing.Size(90, 27);
		this.Subject.TabIndex = 12;
		this.Subject.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75C8FD2D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		this.PWord.AutoSearch = true;
		this.PWord.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.PWord.Indent = 38;
		this.PWord.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468E7DF8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6);
		this.PWord.LabelColor = System.Drawing.Color.Empty;
		this.PWord.Location = new System.Drawing.Point(315, 36);
		this.PWord.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A9214F8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380);
		this.PWord.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PWord.Resetable = true;
		this.PWord.Size = new System.Drawing.Size(190, 28);
		this.PWord.TabIndex = 10;
		this.PKind.AutoHeight = true;
		this.PKind.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.PKind.FlowBreak = false;
		this.PKind.LabelColor = System.Drawing.Color.Empty;
		this.PKind.Location = new System.Drawing.Point(3, 35);
		this.PKind.MaximumSize = new System.Drawing.Size(450, 27);
		this.PKind.MinimumSize = new System.Drawing.Size(310, 0);
		this.PKind.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408BC830 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44);
		this.PKind.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PKind.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x103369E8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
		this.PKind.Size = new System.Drawing.Size(310, 27);
		this.PKind.TabIndex = 26;
		this.btnNew.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.btnNew.FlowBreak = false;
		this.btnNew.Location = new System.Drawing.Point(433, 36);
		this.btnNew.Margin = new System.Windows.Forms.Padding(5, 4, 3, 5);
		this.btnNew.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E465ABF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		this.btnNew.Size = new System.Drawing.Size(72, 25);
		this.btnNew.TabIndex = 10;
		this.btnNew.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14B9A2C1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef);
		this.selMeliCate.Always = true;
		this.selMeliCate.AutoHeight = true;
		this.selMeliCate.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.selMeliCate.Indent = 78;
		this.selMeliCate.IsFinal = true;
		this.selMeliCate.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AC3A98 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611);
		this.selMeliCate.LabelColor = System.Drawing.Color.Empty;
		this.selMeliCate.Location = new System.Drawing.Point(96, 3);
		this.selMeliCate.MaximumSize = new System.Drawing.Size(450, 27);
		this.selMeliCate.MinimumSize = new System.Drawing.Size(310, 0);
		this.selMeliCate.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -629032023 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21);
		this.selMeliCate.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.selMeliCate.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6643446 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8);
		this.selMeliCate.Size = new System.Drawing.Size(409, 27);
		this.selMeliCate.TabIndex = 50;
		this.PType.AutoHeight = true;
		this.PType.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.PType.FlowBreak = false;
		this.PType.Indent = 72;
		this.PType.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x600E97F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2);
		this.PType.LabelColor = System.Drawing.Color.Empty;
		this.PType.Location = new System.Drawing.Point(3, 36);
		this.PType.MaxHeight = -7;
		this.PType.MaximumSize = new System.Drawing.Size(280, 28);
		this.PType.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5110AD9B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
		this.PType.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PType.Size = new System.Drawing.Size(280, 28);
		this.PType.TabIndex = 14;
		this.下载分类属性.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.下载分类属性.FlowBreak = false;
		this.下载分类属性.Location = new System.Drawing.Point(217, 36);
		this.下载分类属性.Margin = new System.Windows.Forms.Padding(5, 4, 3, 5);
		this.下载分类属性.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643C35D8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		this.下载分类属性.Size = new System.Drawing.Size(100, 25);
		this.下载分类属性.TabIndex = 11;
		this.下载分类属性.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DCE9854 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5);
		this.导入产品属性.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.导入产品属性.FlowBreak = false;
		this.导入产品属性.Location = new System.Drawing.Point(325, 36);
		this.导入产品属性.Margin = new System.Windows.Forms.Padding(5, 4, 3, 5);
		this.导入产品属性.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B5CFA4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.导入产品属性.Size = new System.Drawing.Size(100, 25);
		this.导入产品属性.TabIndex = 12;
		this.导入产品属性.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x600E95F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2);
		base.Controls.Add(this.right);
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A95521E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380);
		base.Size = new System.Drawing.Size(1058, 673);
		base.Controls.SetChildIndex(this.right, 0);
		base.Controls.SetChildIndex(base.main, 0);
		base.main.ResumeLayout(false);
		base.tail.ResumeLayout(false);
		base.tail.PerformLayout();
		base.head.ResumeLayout(false);
		base.head.PerformLayout();
		base.ResumeLayout(false);
	}
}
