using System;
using System.ComponentModel;
using System.Drawing;
using System.Globalization;
using System.Windows.Forms;
using Enumer;
using HXK1QWpbr9xpkXuC4o;
using Logic;
using NZsfdnHVTHmTqMuAXn;
using ReoGrid;
using ZYing.Data;
using ZYing.Detail;
using ZYing.Interface;
using ZYing.UI;

namespace ZYing.Channel;

[ToolboxItem(false)]
public class 美客多列表 : CTablePanel
{
	public const string Key = "MeliList";

	public const string Title = "美客多列表";

	private IContainer components;

	private ESearch PWord;

	private ELoader lblTotal;

	private ETurn PTurn;

	private ESelect PLevel;

	private EText PKey;

	private ESelectDateRange PDate;

	private EText txtName;

	private EButton btn查询;

	private ELoader lblNote;

	private EButton btn导出;

	private ECombo PLogin;

	public static 美客多列表 Instance => Global.Create(typeof(美客多列表)) as 美客多列表;

	public 美客多列表()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		Dock = DockStyle.Fill;
		tbl.Columns.AddRange(new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468F7300 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3A5844 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273), 60, 60, HorizontalAlignment.Center, ETheme.Verdana), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16359169 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8), "", 50, 50, hidden: true), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B1B125 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -332272949 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21), 70, 0, ETheme.Verdana), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2E08AE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E428698 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee), 70, 70, ETCellStyle.Number), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D5D17E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FC46BD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5), 100, 0), new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BDC683 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DCA2C1A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5), 100, int.MaxValue), new ETColumn(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3225AC56 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E4263D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c), 50, 0, ETCellStyle.Number)
		{
			Color = Color.Purple
		}, new ETColumn(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A8DBF56 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B3C196 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2), 50, 0, ETCellStyle.Default)
		{
			Color = Color.DarkBlue
		}, new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D5DE8C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B58582 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2), 60, 0), new ETColumn(11, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D35DE2C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0EC06F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd), 100, 0), new ETColumn(9, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6646D1A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -1070091843 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6), 110, 110, ETheme.Verdana)
		{
			Fld = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2004259A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f),
			Order = OrderBy.DESC
		}, new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA77D7D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A9256A8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380), 58, 58, ETCellStyle.OP));
		tbl.OrderChanged = tbl_OrderChanged;
		tbl.RowCreating = tbl_RowCreating;
		tbl.RowSetting = tbl_RowSetting;
		tbl.CellOPClick = tbl_OPClick;
		PLevel.Items = User.Levels;
		PLevel.SelectedChanged += doSelect;
		PLogin.Items = Login.Items;
		PLogin.SelectedChanged += doSelect;
		PKey.OK += doSelect;
		PWord.OK += doSelect;
		PWord.Void += doVoid;
		PWord.RightClick += doSelect;
		PTurn.PageChanged += doFresh;
		btn查询.Click += Btn查询_Click;
		btn导出.Click += Btn导出_Click;
	}

	private async void Btn查询_Click(object sender, EventArgs e)
	{
		string text = txtName.Text;
		if (text.Length == 0)
		{
			lblNote.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37BA8793 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e);
			return;
		}
		ZClient zc = ZClient.Create();
		zc.Loader = lblNote;
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x582CFAE3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd);
		await zc.PostAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -1700618238 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1), text);
		lblNote.Text = zc.Message;
	}

	public override void Load()
	{
		doSelect(this, EventArgs.Empty);
	}

	protected override EControl OnCreateRight(ISource src, string name)
	{
		return new 注册明细(src, name);
	}

	private void tbl_OrderChanged(ETColumn col)
	{
		doSelect(this, EventArgs.Empty);
	}

	protected override void tbl_RowCreating(ETRow tr)
	{
		((ETCellLabel)tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1635721D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8)]).Icon = Icos.Get(IcoEnums.qq);
		((ETCellOP)tr.LastCell).AddRange(new UIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E45B652 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee), IcoEnums.enter, IcoEnums.enter_hover, IcoEnums.enter));
	}

	protected override void tbl_RowSetting(ETRow tr, Record re)
	{
		if (re.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224A62A6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2)))
		{
			tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1165E8FF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca)] = Joint.Encode(re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F216C2A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a)));
		}
		if (re.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6D0CAC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535)))
		{
			tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -1030755062 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947)].Replace(User.Levels);
		}
		if (re.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224C1F96 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2)))
		{
			tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D6E28BE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9)] = From.GetSites(re.GetInts(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x600AAD5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2)));
		}
		if (re.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1705051699 ^ -1969364439 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03)))
		{
			int num = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E33AFD4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6));
			tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x605A69D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2)] = ((num > 0) ? new string('*', num.ToString().Length) : "");
		}
	}

	private void tbl_OPClick(ETCellOP c, UIcon ico)
	{
		ETRow row = c.Row;
		if (ico.Name == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643B4686 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848))
		{
			doJump(row);
		}
	}

	private void doJump(ETRow tr)
	{
		DateTime value = Valid.ToDate(tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11658EF3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca)]);
		if (DateTime.Now.Subtract(value).TotalDays > 180.0)
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7080F54E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834));
		}
		Guid guid = Valid.ToGuid(tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63970D98 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2)]);
		if (guid == Guid.Empty)
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC34B76 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332));
		}
		Global.Start(guid);
	}

	private Para doPara()
	{
		Para para = new Para
		{
			select = true,
			start = PDate.Start.Value,
			end = PDate.End.Value,
			ignore = PWord.RightChecked,
			role = Valid.ToInt(PLevel.Value),
			loginid = Valid.ToInt(PLogin.Value),
			word = PWord.Text,
			order = tbl.OrderString
		};
		string text = PKey.Text;
		if (text.Length > 0)
		{
			para.userid = Valid.ToInt(text);
			if (para.userid == 0)
			{
				para.userid = Joint.GetID(text);
			}
		}
		return para;
	}

	private void doVoid(object sender, EventArgs e)
	{
		PDate.Reset();
		PLevel.Reset();
		PKey.Reset();
		PWord.Reset();
		doSelect(sender, e);
	}

	private void doSelect(object sender, EventArgs e)
	{
		doCommand(doPara());
	}

	private void doFresh(object sender, EventArgs e)
	{
		Para para = doPara();
		para.maxcount = PTurn.MaxCount;
		para.page = PTurn.Page;
		doCommand(para);
	}

	private async void doCommand(Para p)
	{
		ZClient zc = ZClient.Create();
		zc.Loader = lblTotal;
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86C79DF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		await zc.PostAsync(p);
		Record record = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACE6804 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057)];
		lblTotal.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527F13EB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1) + record.MaxCount + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D3830A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda);
		tbl.DataSource = record;
		PTurn.Read(record);
	}

	private void PWord_OK(object sender, EventArgs e)
	{
		tbl.Reset();
		tbl.Like(PWord.Text.Trim());
	}

	private async void Btn导出_Click(object sender, EventArgs e)
	{
		ZClient zc = ZClient.Create();
		zc.UI = lblTotal;
		zc.Button = btn导出;
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x502DB20F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841);
		Para p = doPara();
		await zc.PostAsync(p);
		zc.Throw();
		Record record = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C08DCA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03)];
		if (record.Count == 0)
		{
			EDialog.Show(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C965547 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A8ED1E6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21));
		}
		else
		{
			下载Excel(record);
		}
	}

	public static void 下载Excel(Record re)
	{
		string[] array = new string[13]
		{
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC34B52 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68ABD556 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x511614BD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3C8260 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5279A5A9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CBD695 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1705051699 ^ -1536664901 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2B68D4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C125A0A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D09BC2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D6F4074 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x191EECE8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643C6C22 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848)
		};
		SaveFileDialog saveXlsDialog = Global.SaveXlsDialog;
		saveXlsDialog.FileName = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63916582 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2) + DateTime.Now.ToString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C11466 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03), DateTimeFormatInfo.InvariantInfo);
		if (saveXlsDialog.ShowDialog() != DialogResult.OK)
		{
			return;
		}
		string fileName = saveXlsDialog.FileName;
		Workbook workbook = new Workbook();
		Worksheet worksheet = workbook.InsertWorksheet(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B19396F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791));
		worksheet.ColumnCount = array.Length;
		for (int i = 0; i < array.Length; i++)
		{
			int num = 100;
			string text = array[i];
			int num2 = text.IndexOf('|');
			if (num2 > -1)
			{
				num = Valid.ToInt(text.Substring(num2 + 1));
				text = (array[i] = text.Substring(0, num2));
			}
			if (num > 0)
			{
				worksheet.ColumnHeaders[i].Width = (ushort)num;
			}
			Cell cell = worksheet.Cells[0, i];
			cell.ApplyHeadStyle();
			cell.Data = text;
		}
		worksheet.RowCount = re.Count + 1;
		Color back = Color.Empty;
		int num3 = 0;
		int num4 = 0;
		while (re.Read())
		{
			int num5 = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DC8B2D4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5));
			if (num5 != num4)
			{
				num4 = num5;
				back = ((num3 % 2 != 0) ? Color.FromArgb(210 + Tool.GetRandomNext(45), 210 + Tool.GetRandomNext(45), 210 + Tool.GetRandomNext(45)) : Color.Empty);
				num3++;
			}
			Color color = EColors.GetColor(re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B276CB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552)));
			for (int j = 0; j < array.Length; j++)
			{
				Cell cell2 = worksheet.Cells[re.Cursor + 1, j];
				cell2.ApplyBackColor(back);
				cell2.Style.TextColor = color;
				string text2 = array[j];
				switch (<PrivateImplementationDetails>.ComputeStringHash(text2))
				{
				case 2246226042u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3A5844 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273)))
					{
						break;
					}
					cell2.Data = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-2066591320) ^ 0x717594DB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1));
					continue;
				case 2270967379u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468EF228 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6)))
					{
						break;
					}
					cell2.Data = Joint.Encode(re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DC8A6DE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5)));
					continue;
				case 2728616837u:
				{
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86095C3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788)))
					{
						break;
					}
					int num6 = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x58257151 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd));
					cell2.Data = ((num6 > 0) ? num6.ToString() : "");
					continue;
				}
				case 4271031382u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D335CC6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b)))
					{
						break;
					}
					cell2.Data = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E47C69A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee));
					continue;
				case 4111700684u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C987D89 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947)))
					{
						break;
					}
					cell2.Data = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68ACA174 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611));
					continue;
				case 2646960156u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643C7744 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848)))
					{
						break;
					}
					cell2.Data = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA51DFD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97));
					continue;
				case 1072339432u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E428786 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee)))
					{
						break;
					}
					cell2.Data = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA582123 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1));
					continue;
				case 846798487u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E3C384 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce)))
					{
						break;
					}
					cell2.Data = User.GetLevel(re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A39C4EC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273)));
					continue;
				case 633585526u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -565996370 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689)))
					{
						break;
					}
					cell2.Data = From.GetSites(re.GetInts(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408DCE3A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44)));
					continue;
				case 2080157015u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51161E95 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7)))
					{
						break;
					}
					cell2.Data = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5027D1E9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841));
					continue;
				case 1778475639u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -1222348900 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f)))
					{
						break;
					}
					cell2.Data = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68D07EEE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6));
					continue;
				case 2670303808u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A924DFC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380)))
					{
						break;
					}
					cell2.Data = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E53174 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce));
					continue;
				case 1166687027u:
					if (!(text2 == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2B726C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3)))
					{
						break;
					}
					cell2.Data = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E453D24 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee));
					continue;
				}
				cell2.Data = "";
			}
		}
		workbook.Save(fileName);
		EDialog.Open(fileName);
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
		this.PWord = new ZYing.UI.ESearch();
		this.lblTotal = new ZYing.UI.ELoader();
		this.PTurn = new ZYing.UI.ETurn();
		this.PLevel = new ZYing.UI.ESelect();
		this.PKey = new ZYing.UI.EText();
		this.PDate = new ZYing.UI.ESelectDateRange();
		this.txtName = new ZYing.UI.EText();
		this.btn查询 = new ZYing.UI.EButton();
		this.lblNote = new ZYing.UI.ELoader();
		this.btn导出 = new ZYing.UI.EButton();
		this.PLogin = new ZYing.UI.ECombo();
		base.main.SuspendLayout();
		base.tail.SuspendLayout();
		base.head.SuspendLayout();
		base.SuspendLayout();
		base.right.Location = new System.Drawing.Point(1386, 0);
		base.right.MaximumSize = new System.Drawing.Size(600, 500000);
		base.right.MinimumSize = new System.Drawing.Size(440, 0);
		base.right.Size = new System.Drawing.Size(440, 673);
		base.right.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E47C7E6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee);
		base.main.Size = new System.Drawing.Size(1386, 673);
		base.tail.Controls.Add(this.PTurn);
		base.tail.Controls.Add(this.lblTotal);
		base.tail.Controls.Add(this.btn导出);
		base.tail.Controls.Add(this.txtName);
		base.tail.Controls.Add(this.btn查询);
		base.tail.Controls.Add(this.lblNote);
		base.tail.Location = new System.Drawing.Point(7, 564);
		base.tail.Size = new System.Drawing.Size(1371, 81);
		base.tbl.Location = new System.Drawing.Point(10, 43);
		base.tbl.MaxHeight = 539;
		base.tbl.Size = new System.Drawing.Size(1365, 518);
		base.tbl.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408DC4DE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44);
		base.head.Controls.Add(this.PDate);
		base.head.Controls.Add(this.PLevel);
		base.head.Controls.Add(this.PLogin);
		base.head.Controls.Add(this.PKey);
		base.head.Controls.Add(this.PWord);
		base.head.Size = new System.Drawing.Size(1371, 34);
		this.PWord.FlowBreak = false;
		this.PWord.FlowWrap = false;
		this.PWord.Indent = 42;
		this.PWord.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468FDCA6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6);
		this.PWord.Location = new System.Drawing.Point(701, 3);
		this.PWord.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16352B47 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8);
		this.PWord.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PWord.Resetable = true;
		this.PWord.RightCheckable = true;
		this.PWord.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x582010DB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd);
		this.PWord.Size = new System.Drawing.Size(321, 28);
		this.PWord.TabIndex = 7;
		this.lblTotal.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.lblTotal.ForeColor = System.Drawing.Color.Black;
		this.lblTotal.Location = new System.Drawing.Point(1238, 3);
		this.lblTotal.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5826C31D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd);
		this.lblTotal.Padding = new System.Windows.Forms.Padding(5, 3, 2, 3);
		this.lblTotal.Size = new System.Drawing.Size(129, 26);
		this.lblTotal.TabIndex = 10;
		this.lblTotal.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7646F1A5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
		this.lblTotal.TextAlign = System.Windows.Forms.HorizontalAlignment.Right;
		this.PTurn.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.PTurn.FlowBreak = false;
		this.PTurn.Location = new System.Drawing.Point(3, 5);
		this.PTurn.Margin = new System.Windows.Forms.Padding(3, 5, 3, 10);
		this.PTurn.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3B93F6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273);
		this.PTurn.Padding = new System.Windows.Forms.Padding(4);
		this.PTurn.Size = new System.Drawing.Size(1229, 27);
		this.PTurn.TabIndex = 12;
		this.PTurn.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A5451E2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39);
		this.PLevel.AutoHeight = true;
		this.PLevel.FlowBreak = false;
		this.PLevel.Indent = 42;
		this.PLevel.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F27101E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a);
		this.PLevel.Location = new System.Drawing.Point(273, 3);
		this.PLevel.MaximumSize = new System.Drawing.Size(220, 28);
		this.PLevel.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DCEDAFE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5);
		this.PLevel.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PLevel.Size = new System.Drawing.Size(130, 28);
		this.PLevel.TabIndex = 12;
		this.PKey.FlowBreak = false;
		this.PKey.Indent = 42;
		this.PKey.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE4A622 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde);
		this.PKey.Location = new System.Drawing.Point(575, 3);
		this.PKey.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACF6FE6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057);
		this.PKey.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PKey.Size = new System.Drawing.Size(120, 28);
		this.PKey.TabIndex = 13;
		this.PKey.ValidType = ZYing.UI.EValidType.EN;
		this.PDate.FlowBreak = false;
		this.PDate.Indent = 42;
		this.PDate.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A39C5F6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273);
		this.PDate.Location = new System.Drawing.Point(3, 3);
		this.PDate.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9CBB81 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
		this.PDate.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PDate.Size = new System.Drawing.Size(264, 28);
		this.PDate.TabIndex = 11;
		this.txtName.FlowBreak = false;
		this.txtName.Indent = 50;
		this.txtName.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D3549C8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
		this.txtName.Location = new System.Drawing.Point(81, 45);
		this.txtName.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408B3526 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44);
		this.txtName.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtName.Size = new System.Drawing.Size(263, 28);
		this.txtName.TabIndex = 11;
		this.btn查询.FlowBreak = false;
		this.btn查询.Location = new System.Drawing.Point(350, 45);
		this.btn查询.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9C4917 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
		this.btn查询.Padding = new System.Windows.Forms.Padding(8, 4, 8, 4);
		this.btn查询.Size = new System.Drawing.Size(86, 27);
		this.btn查询.TabIndex = 12;
		this.btn查询.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B4F51E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53);
		this.lblNote.AutoHeight = true;
		this.lblNote.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.lblNote.Location = new System.Drawing.Point(442, 45);
		this.lblNote.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FAB823 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5);
		this.lblNote.Size = new System.Drawing.Size(925, 28);
		this.lblNote.TabIndex = 13;
		this.btn导出.FlowBreak = false;
		this.btn导出.ForeColor = System.Drawing.Color.DarkBlue;
		this.btn导出.Location = new System.Drawing.Point(3, 48);
		this.btn导出.Margin = new System.Windows.Forms.Padding(3, 6, 3, 8);
		this.btn导出.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10345A6C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
		this.btn导出.Size = new System.Drawing.Size(72, 25);
		this.btn导出.TabIndex = 13;
		this.btn导出.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA09E65F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.PLogin.AutoHeight = true;
		this.PLogin.FlowBreak = false;
		this.PLogin.Indent = 42;
		this.PLogin.IsExact = true;
		this.PLogin.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA09E66B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.PLogin.Location = new System.Drawing.Point(409, 3);
		this.PLogin.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F1B17F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1);
		this.PLogin.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PLogin.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6E5F2E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.PLogin.Size = new System.Drawing.Size(160, 28);
		this.PLogin.TabIndex = 243;
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DC8C6F2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5);
		base.Size = new System.Drawing.Size(1826, 673);
		base.main.ResumeLayout(false);
		base.tail.ResumeLayout(false);
		base.tail.PerformLayout();
		base.head.ResumeLayout(false);
		base.ResumeLayout(false);
	}
}
