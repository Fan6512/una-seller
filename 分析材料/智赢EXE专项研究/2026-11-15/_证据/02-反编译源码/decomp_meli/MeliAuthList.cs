using System;
using System.Collections.Specialized;
using System.ComponentModel;
using System.Drawing;
using System.Runtime.CompilerServices;
using System.Threading.Tasks;
using System.Windows.Forms;
using Enumer;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Data;
using ZYing.Detail;
using ZYing.Dialog;
using ZYing.Interface;
using ZYing.Meli;
using ZYing.UI;

namespace ZYing.Channel;

[ToolboxItem(false)]
public class Meli授权列表 : CTablePanelScroll
{
	[CompilerGenerated]
	private sealed class <>c__DisplayClass10_0
	{
		public ZClient zc;

		public ETCellOP c;

		public Meli授权列表 <>4__this;

		public <>c__DisplayClass10_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}

		internal void <Tbl_OPClick>b__1(ZClient <p0>)
		{
			EDialog.Show(zc.Message, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E30C164 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6));
		}

		internal void <Tbl_OPClick>b__3(ZClient <p0>)
		{
			ETRow row = c.Row;
			<>4__this.tbl.Rows.Remove(row);
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass9_0
	{
		public ZClient zc;

		public ETCellCheck c;

		public Meli授权列表 <>4__this;

		public <>c__DisplayClass9_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}

		internal void <Tbl_CellChecked>b__0(Exception <p0>)
		{
			EDialog.Show(zc.Message, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16359B5D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8));
			c.Checked = !c.Checked;
		}

		internal void <Tbl_CellChecked>b__1(ZClient <p0>)
		{
			c.Row.Color = (c.Checked ? EColors.GetColor(c.Row[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F41692A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835)]) : Color.FromArgb(190, 190, 190));
			<>4__this.Load();
		}
	}

	public const string Key = "Acc";

	public const string Caption = "账户授权";

	private IContainer components;

	private ESearch PWord;

	private EButton btnNew;

	private ELoader lblTotal;

	private ELabel Subject;

	private ECombo PLogin;

	private ESelect pType;

	private EFilter selFlags;

	private ELine eLine1;

	private EFilter selAreas;

	private ETurn PTurn;

	private ELabel lbl收款;

	private EButton btn应用;

	private EButton btn搜索;

	private ECombo sel归属;

	private ELabel lbTip;

	public static Meli授权列表 Instance => Global.Create(typeof(Meli授权列表)) as Meli授权列表;

	public Meli授权列表()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		Subject.Icon = (right.Icon = Icos.Get(IcoEnums.article));
		lbl收款.Url = Enumer.Web.WorldFirst;
		tbl.Columns.AddRange(new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A32948 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2DBB5A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3), 45, 0, hidden: true), new ETColumn(11, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB68AEB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BD6F2B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef), 45, 0, hidden: true), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE4F72E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB0B6CB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018), 60, hidden: true), new ETColumn(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E4B9C2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0AF48B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd), 70, 0), new ETColumn(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60120EF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B26448 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53), 40, 0), new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51107D09 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0E8A03 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd), 45, 45, ETCellStyle.CheckBox), new ETColumn(6, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB04F05 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x20028286 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f), 60), new ETColumn(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACE0956 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -2110798095 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594), 60), new ETColumn(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA6DA05 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5112E3F1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7), 60), new ETColumn(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC1442A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -408819306 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834), 60), new ETColumn(11, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA583295 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x1C00186F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1), 60, 60), new ETColumn(11, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527EC391 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE11E6E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde), 60, 60), new ETColumn(14, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CD3A9D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x66521D6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8), 60, 60), new ETColumn(15, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x60C6F39 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F269B0A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a), 100, 200), new ETColumn(13, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51155771 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x50109DB8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e), 70, int.MaxValue), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D6F1802 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F40254C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835), 35, 60, ETCellStyle.OP));
		tbl.RowCreating = tbl_RowCreating;
		tbl.RowSetting = tbl_RowSetting;
		tbl.CellOPClick = Tbl_OPClick;
		tbl.CellChecked = Tbl_CellChecked;
		pType.SelectedChanged += DoSelect;
		pType.Items = new EItemCollection
		{
			new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -811429991 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A8F6EC6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21)),
			new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E4C6E3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1051B9AE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf)),
			new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x20029314 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AA5F92 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611)),
			new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6061C49 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2005AFDE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f)),
			new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A3A4D0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643BABB2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848))
		};
		selFlags.Items.AddRange(new EItem("", kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527F7117 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1)), new EItem(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B10199 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e), Color.DarkGreen), new EItem(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51155A0D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7), Color.DarkRed), new EItem(9, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA9B9EC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b), Color.DarkGray));
		selFlags.SelectedChanged += DoFlags;
		selAreas.Items = Enumer.Meli.Abbrs.Copy();
		selAreas.Items.RemoveAt(0);
		selAreas.Items.Insert(0, new EItem("", kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F163E1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1)));
		selAreas.SelectedChanged += DoAreas;
		PTurn.PageChanged += DoFresh;
		if (!Member.只看自己店铺 && (Member.Limit.系统.看业务员 || Member.Limit.系统.改业务员))
		{
			PLogin.Items = Login.Items;
			PLogin.Visible = PLogin.Items.VisibleCount > 1;
			PLogin.SelectedChanged += DoSelect;
			if (Member.UserID == 6238)
			{
				sel归属.Items = Login.Items;
				sel归属.Visible = sel归属.Items.VisibleCount > 1;
				sel归属.SelectedChanged += DoSelect;
			}
		}
		PWord.OK += DoSelect;
		PWord.Void += DoVoid;
		btnNew.Click += New_Click;
		btn应用.Visible = Member.Is系统管理员;
		btn应用.Click += delegate
		{
			美客多应用Dialog.Instance.Show(this);
		};
		MeliAppInfo.Load().GetAwaiter();
		btn搜索.Visible = Member.Is系统管理员 && Member.UserID.In(1000, 1801, 1051);
		btn搜索.Click += delegate
		{
			搜索店铺Dialog.Instance.Show(this);
		};
	}

	public override void Load()
	{
		DoFlags(this, EventArgs.Empty);
	}

	protected override EControl OnCreateRight(ISource src, string name)
	{
		return new Meli授权详情(src, name)
		{
			updated = delegate
			{
				right.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DCCCC4C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5);
				right.ScrollTop();
				DoSelect(this, EventArgs.Empty);
			}
		};
	}

	protected override void tbl_RowCreating(ETRow tr)
	{
		if (Member.Is系统管理员)
		{
			tr.LastCell.AddRange(new UIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1035F82E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d)));
		}
	}

	protected override void tbl_RowSetting(ETRow tr, Record re)
	{
		if (re.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA58E32B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1)))
		{
			string key = re[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A926282 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380)].ToString();
			string text = re[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5110EF31 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7)].ToString();
			ETCell eTCell = tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7647509D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835)];
			eTCell.Icon = World.GetIcon(Enumer.Meli.GetArea(key));
			string text2 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5826CC7D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd) + Enumer.Meli.Items.FindKey(key)?.Text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x8669A0D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
			if (re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2CB5C0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3)))
			{
				text2 += kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB199FF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018);
			}
			eTCell.Text = (text.IsNullOrEmpty() ? (re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC5400E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332)) + text2) : (text + text2));
		}
		if (re.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68ADBCF2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611)))
		{
			int num = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E46DABD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689));
			bool flag = num.In(1, 2);
			tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5010E8E8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e)].Text = (flag ? kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0F83EF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd) : kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3A8A1E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273));
			tr.Color = (flag ? EColors.GetColor(tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A288AC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010)]) : Color.FromArgb(190, 190, 190));
			tr.BackColor = ((num == 2) ? Color.FromArgb(255, 220, 220) : Color.Empty);
		}
		if (re.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FA26D5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5)))
		{
			tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A97856E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380)] = Login.GetName(re.GetInts(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x103543AC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d)));
		}
		if (re.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224AF304 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2)))
		{
			double num2 = re[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CD3A9D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206)].ToDouble();
			tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5027B6C3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841)] = ((num2 > 0.0) ? num2.ToString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E465FB7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689)) : "");
		}
		bool flag2 = re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F20CCD0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a));
		re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527E78CD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1));
		int num3 = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68ADB314 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611));
		ETCell eTCell2 = tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1163BF2B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca)];
		eTCell2.Text = Valid.GetEnumDescription((MeliShop.ShopType)num3);
		tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB0DC31 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4)] = "";
		if (flag2)
		{
			eTCell2.Color = tr.Color;
			string appId = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3D016C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273));
			tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68D70F8E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6)] = MeliAppInfo.GetByAppId(appId)?.AppName ?? "";
		}
		else
		{
			eTCell2.Text = Enumer.Meli.Items.FindKey(re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2002813A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f)))?.Text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -265259263 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
			eTCell2.Color = Color.Red;
		}
	}

	private async void Tbl_CellChecked(ETCellCheck c)
	{
		<>c__DisplayClass9_0 CS$<>8__locals20 = new <>c__DisplayClass9_0();
		CS$<>8__locals20.c = c;
		CS$<>8__locals20.<>4__this = this;
		if (CS$<>8__locals20.c.Checked)
		{
			if (!EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60F07E7F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5)))
			{
				CS$<>8__locals20.c.Checked = false;
				return;
			}
		}
		else if (!EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A31CA32 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273)))
		{
			CS$<>8__locals20.c.Checked = true;
			return;
		}
		int num = (CS$<>8__locals20.c.Checked ? 1 : 9);
		NameValueCollection nv = new NameValueCollection
		{
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0F9A11 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd)] = CS$<>8__locals20.c.Row[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x32244238 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594)],
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9CD4F1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947)] = num.ToString()
		};
		CS$<>8__locals20.zc = ZClient.Create(ZClient.美客多);
		CS$<>8__locals20.zc.UI = this;
		CS$<>8__locals20.zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA395A0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b);
		CS$<>8__locals20.zc.Failed = delegate
		{
			EDialog.Show(CS$<>8__locals20.zc.Message, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16359B5D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8));
			CS$<>8__locals20.c.Checked = !CS$<>8__locals20.c.Checked;
		};
		CS$<>8__locals20.zc.Finish = delegate
		{
			CS$<>8__locals20.c.Row.Color = (CS$<>8__locals20.c.Checked ? EColors.GetColor(CS$<>8__locals20.c.Row[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F41692A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835)]) : Color.FromArgb(190, 190, 190));
			CS$<>8__locals20.<>4__this.Load();
		};
		await CS$<>8__locals20.zc.PostAsync(nv);
	}

	private async void Tbl_OPClick(ETCellOP c, UIcon ico)
	{
		<>c__DisplayClass10_0 CS$<>8__locals21 = new <>c__DisplayClass10_0();
		CS$<>8__locals21.c = c;
		CS$<>8__locals21.<>4__this = this;
		string name = ico.Name;
		if (!(name == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B30FFC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2)))
		{
			if (name == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x708627E8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834))
			{
				if (CS$<>8__locals21.c.Row[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB73F33 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4)] == 1.ToString())
				{
					EDialog.Show(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x511B5597 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x665B470 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8));
				}
				else if (EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46844FC2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6)) && EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68A697C8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611)))
				{
					CS$<>8__locals21.zc = ZClient.Create(ZClient.美客多);
					CS$<>8__locals21.zc.UI = this;
					CS$<>8__locals21.zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D3E17DE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
					CS$<>8__locals21.zc.Failed = (Exception ex) =>
					{
						EDialog.Show(ex);
					};
					CS$<>8__locals21.zc.Finish = delegate
					{
						ETRow row = CS$<>8__locals21.c.Row;
						CS$<>8__locals21.<>4__this.tbl.Rows.Remove(row);
					};
					await CS$<>8__locals21.zc.PostAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76465E47 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835), CS$<>8__locals21.c.Row.Key);
				}
			}
			else
			{
				tbl.CurrentRow = CS$<>8__locals21.c.Row;
			}
		}
		else if (EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x40808F6E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44)))
		{
			CS$<>8__locals21.zc = ZClient.Create(ZClient.美客多);
			CS$<>8__locals21.zc.UI = this;
			CS$<>8__locals21.zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA04A695 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
			CS$<>8__locals21.zc.Failed = (Exception ex) =>
			{
				EDialog.Show(ex);
			};
			CS$<>8__locals21.zc.Finish = delegate
			{
				EDialog.Show(CS$<>8__locals21.zc.Message, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E30C164 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6));
			};
			await CS$<>8__locals21.zc.PostAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA8A9E6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b), CS$<>8__locals21.c.Row[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5010FCD4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e)]);
		}
	}

	private async void DoFresh(object sender, EventArgs e)
	{
		Para para = DoPara();
		para.maxcount = PTurn.MaxCount;
		para.page = PTurn.Page;
		await DoCommand(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB2C38B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018), para);
	}

	private void DoVoid(object sender, EventArgs e)
	{
		head.Reset();
		DoFlags(sender, e);
	}

	private async Task DoCommand(string cmd, Para p)
	{
		ZClient zc = ZClient.Create(ZClient.美客多);
		zc.UI = this;
		zc.Loader = lblTotal;
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x764C4BF7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835) + cmd;
		await zc.PostAsync(p);
		zc.Throw();
		if (zc.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B53B23 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552)))
		{
			selFlags.Read(zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EAE45EA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b)]);
		}
		if (zc.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B1E977 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552)))
		{
			selAreas.Read(zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D4D23A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda)]);
		}
		Record record = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10510A44 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf)];
		tbl.DataSource = record;
		PTurn.Read(record);
		lblTotal.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACE44A8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057) + record.MaxCount + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14B4BCF9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef);
	}

	private async void DoFlags(object sender, EventArgs e)
	{
		await DoCommand(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -11378730 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2), DoPara());
	}

	private async void DoAreas(object sender, EventArgs e)
	{
		await DoCommand(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2F2B44 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3), DoPara());
	}

	private async void DoSelect(object sender, EventArgs e)
	{
		await DoCommand(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468D89B8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6), DoPara());
	}

	private Para DoPara()
	{
		return new Para
		{
			loginid = PLogin.Value.ToInt(),
			role = sel归属.Value.ToInt(),
			key = PWord.Text,
			flag = selFlags.Value.ToInt(),
			abbr = selAreas.Value,
			pagesize = 50,
			type = pType.Value.ToInt()
		};
	}

	private void New_Click(object sender, EventArgs e)
	{
		Meli授权详情 page = new Meli授权详情(tbl, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E462BB2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee))
		{
			Inserted = delegate
			{
				right.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x862824F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
				right.ScrollTop();
				DoSelect(this, EventArgs.Empty);
			}
		};
		right.Page = page;
		right.Percent = right.Width * 100 / Width;
		right.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2AE284 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3);
		tbl.CurrentRow = null;
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
		this.Subject = new ZYing.UI.ELabel();
		this.PWord = new ZYing.UI.ESearch();
		this.btnNew = new ZYing.UI.EButton();
		this.lblTotal = new ZYing.UI.ELoader();
		this.PLogin = new ZYing.UI.ECombo();
		this.pType = new ZYing.UI.ESelect();
		this.selFlags = new ZYing.UI.EFilter();
		this.eLine1 = new ZYing.UI.ELine();
		this.selAreas = new ZYing.UI.EFilter();
		this.PTurn = new ZYing.UI.ETurn();
		this.lbl收款 = new ZYing.UI.ELabel();
		this.btn应用 = new ZYing.UI.EButton();
		this.btn搜索 = new ZYing.UI.EButton();
		this.sel归属 = new ZYing.UI.ECombo();
		this.lbTip = new ZYing.UI.ELabel();
		base.main.SuspendLayout();
		base.tail.SuspendLayout();
		base.head.SuspendLayout();
		base.SuspendLayout();
		base.right.Location = new System.Drawing.Point(887, 0);
		base.right.MaximumSize = new System.Drawing.Size(800, 500000);
		base.right.MinimumSize = new System.Drawing.Size(450, 0);
		base.right.Size = new System.Drawing.Size(450, 568);
		base.right.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14B9FC77 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef);
		base.main.Size = new System.Drawing.Size(887, 568);
		base.tail.Controls.Add(this.PTurn);
		base.tail.Controls.Add(this.lblTotal);
		base.tail.Controls.Add(this.btnNew);
		base.tail.Controls.Add(this.btn应用);
		base.tail.Controls.Add(this.btn搜索);
		base.tail.Controls.Add(this.lbTip);
		base.tail.Controls.Add(this.lbl收款);
		base.tail.Location = new System.Drawing.Point(7, 482);
		base.tail.Size = new System.Drawing.Size(872, 69);
		base.tbl.Location = new System.Drawing.Point(10, 109);
		base.tbl.MaxHeight = 380;
		base.tbl.Size = new System.Drawing.Size(866, 370);
		base.tbl.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A978854 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380);
		base.head.Controls.Add(this.Subject);
		base.head.Controls.Add(this.pType);
		base.head.Controls.Add(this.PLogin);
		base.head.Controls.Add(this.sel归属);
		base.head.Controls.Add(this.PWord);
		base.head.Controls.Add(this.eLine1);
		base.head.Controls.Add(this.selFlags);
		base.head.Controls.Add(this.selAreas);
		base.head.Size = new System.Drawing.Size(872, 100);
		this.Subject.AutoSize = false;
		this.Subject.FlowBreak = false;
		this.Subject.Fonts = ZYing.UI.EFonts.微软雅黑Bold;
		this.Subject.Location = new System.Drawing.Point(3, 3);
		this.Subject.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E30712E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.Subject.Padding = new System.Windows.Forms.Padding(0, 5, 6, 3);
		this.Subject.Size = new System.Drawing.Size(126, 27);
		this.Subject.TabIndex = 11;
		this.Subject.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x582351E5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd);
		this.PWord.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.PWord.Indent = 42;
		this.PWord.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B35EB8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.PWord.Location = new System.Drawing.Point(535, 3);
		this.PWord.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E466153 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		this.PWord.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PWord.Resetable = true;
		this.PWord.Size = new System.Drawing.Size(333, 28);
		this.PWord.TabIndex = 7;
		this.btnNew.FlowBreak = false;
		this.btnNew.ForeColor = System.Drawing.Color.Black;
		this.btnNew.Location = new System.Drawing.Point(3, 40);
		this.btnNew.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B34311 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		this.btnNew.Padding = new System.Windows.Forms.Padding(10, 3, 10, 3);
		this.btnNew.Size = new System.Drawing.Size(90, 25);
		this.btnNew.TabIndex = 7;
		this.btnNew.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C111942 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.lblTotal.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.lblTotal.ForeColor = System.Drawing.Color.Black;
		this.lblTotal.Location = new System.Drawing.Point(706, 7);
		this.lblTotal.Margin = new System.Windows.Forms.Padding(3, 7, 3, 3);
		this.lblTotal.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2002F76C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f);
		this.lblTotal.Padding = new System.Windows.Forms.Padding(5, 3, 2, 3);
		this.lblTotal.Size = new System.Drawing.Size(162, 26);
		this.lblTotal.TabIndex = 10;
		this.lblTotal.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F40672C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835);
		this.lblTotal.TextAlign = System.Windows.Forms.HorizontalAlignment.Right;
		this.PLogin.Always = true;
		this.PLogin.AutoHeight = true;
		this.PLogin.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.PLogin.FlowBreak = false;
		this.PLogin.IsExact = true;
		this.PLogin.Location = new System.Drawing.Point(187, 3);
		this.PLogin.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1705051699 ^ -1959081468 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca);
		this.PLogin.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PLogin.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB0585F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018);
		this.PLogin.Size = new System.Drawing.Size(180, 28);
		this.PLogin.TabIndex = 290;
		this.PLogin.Visible = false;
		this.pType.AutoHeight = true;
		this.pType.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.pType.FlowBreak = false;
		this.pType.Indent = 42;
		this.pType.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0AB915 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.pType.Location = new System.Drawing.Point(-5, 3);
		this.pType.MaximumSize = new System.Drawing.Size(280, 28);
		this.pType.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B57EC2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53);
		this.pType.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.pType.Size = new System.Drawing.Size(186, 28);
		this.pType.TabIndex = 14;
		this.selFlags.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.selFlags.Indent = 50;
		this.selFlags.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527EF3B7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1);
		this.selFlags.Location = new System.Drawing.Point(3, 42);
		this.selFlags.Margin = new System.Windows.Forms.Padding(3, 1, 3, 3);
		this.selFlags.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E41B536 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee);
		this.selFlags.Padding = new System.Windows.Forms.Padding(0, 3, 6, 3);
		this.selFlags.Size = new System.Drawing.Size(865, 27);
		this.selFlags.TabIndex = 16;
		this.eLine1.Location = new System.Drawing.Point(3, 36);
		this.eLine1.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -912010944 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		this.eLine1.Size = new System.Drawing.Size(865, 2);
		this.selAreas.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.selAreas.Indent = 50;
		this.selAreas.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B0D016 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53);
		this.selAreas.Location = new System.Drawing.Point(3, 72);
		this.selAreas.Margin = new System.Windows.Forms.Padding(3, 0, 3, 3);
		this.selAreas.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B1FCA87 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791);
		this.selAreas.Padding = new System.Windows.Forms.Padding(0, 0, 6, 3);
		this.selAreas.Size = new System.Drawing.Size(865, 25);
		this.selAreas.TabIndex = 15;
		this.selAreas.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68A89ADE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611);
		this.PTurn.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.PTurn.FlowBreak = false;
		this.PTurn.Location = new System.Drawing.Point(3, 5);
		this.PTurn.Margin = new System.Windows.Forms.Padding(3, 5, 3, 5);
		this.PTurn.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0EFFE5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.PTurn.Padding = new System.Windows.Forms.Padding(4);
		this.PTurn.Size = new System.Drawing.Size(697, 27);
		this.PTurn.TabIndex = 14;
		this.PTurn.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA6B65D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
		this.lbl收款.ForeColor = System.Drawing.Color.DarkBlue;
		this.lbl收款.Location = new System.Drawing.Point(702, 40);
		this.lbl收款.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x32203368 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		this.lbl收款.Padding = new System.Windows.Forms.Padding(6, 3, 6, 3);
		this.lbl收款.Size = new System.Drawing.Size(153, 26);
		this.lbl收款.TabIndex = 457;
		this.lbl收款.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x8611F67 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		this.btn应用.FlowBreak = false;
		this.btn应用.ForeColor = System.Drawing.Color.Black;
		this.btn应用.Location = new System.Drawing.Point(99, 40);
		this.btn应用.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B42089 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		this.btn应用.Padding = new System.Windows.Forms.Padding(10, 3, 10, 3);
		this.btn应用.Size = new System.Drawing.Size(76, 25);
		this.btn应用.TabIndex = 11;
		this.btn应用.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9B3719 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
		this.btn应用.Visible = false;
		this.btn搜索.FlowBreak = false;
		this.btn搜索.ForeColor = System.Drawing.Color.Black;
		this.btn搜索.Location = new System.Drawing.Point(181, 40);
		this.btn搜索.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC26566 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332);
		this.btn搜索.Padding = new System.Windows.Forms.Padding(10, 3, 10, 3);
		this.btn搜索.Size = new System.Drawing.Size(76, 25);
		this.btn搜索.TabIndex = 12;
		this.btn搜索.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AA5F0E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611);
		this.btn搜索.Visible = false;
		this.sel归属.Always = true;
		this.sel归属.AutoHeight = true;
		this.sel归属.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.sel归属.FlowBreak = false;
		this.sel归属.IsExact = true;
		this.sel归属.Location = new System.Drawing.Point(373, 3);
		this.sel归属.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E41397D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		this.sel归属.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.sel归属.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA091DDB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.sel归属.Size = new System.Drawing.Size(156, 28);
		this.sel归属.TabIndex = 291;
		this.sel归属.Visible = false;
		this.lbTip.Blink = true;
		this.lbTip.FlowBreak = false;
		this.lbTip.Fonts = ZYing.UI.EFonts.微软雅黑Bold;
		this.lbTip.ForeColor = System.Drawing.Color.Tomato;
		this.lbTip.Location = new System.Drawing.Point(263, 40);
		this.lbTip.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3A821A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273);
		this.lbTip.Padding = new System.Windows.Forms.Padding(6, 3, 6, 3);
		this.lbTip.Size = new System.Drawing.Size(433, 25);
		this.lbTip.TabIndex = 458;
		this.lbTip.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA5FCD69 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1);
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10561B38 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf);
		base.Size = new System.Drawing.Size(1337, 568);
		base.main.ResumeLayout(false);
		base.tail.ResumeLayout(false);
		base.tail.PerformLayout();
		base.head.ResumeLayout(false);
		base.ResumeLayout(false);
	}
}
