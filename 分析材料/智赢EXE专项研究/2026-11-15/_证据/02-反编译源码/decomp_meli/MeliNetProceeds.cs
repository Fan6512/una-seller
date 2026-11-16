using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Drawing;
using System.Drawing.Drawing2D;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Forms;
using Enumer;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Crawl;
using ZYing.Data;
using ZYing.UI;

namespace ZYing.Channel;

[ToolboxItem(false)]
public class 任务Meli净收益估算 : EContainer
{
	private class SaleInfo
	{
		public int from;

		public int id;

		public string thumb;

		public int siteid;

		public string source;

		public int weight;

		public int[] size;

		public string cur;

		public decimal cost;

		public decimal netproceed;

		public int localid;

		public int loginid;

		public DateTime ptime;

		public bool? Successed { get; set; }

		public string Message { get; set; }

		public string DetailMessage { get; set; }

		public SaleInfo()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}
	}

	public const string Key = "meli_net_proceeds_estimate";

	public const string Caption = "净收益估算";

	private bool _busing;

	private IContainer components;

	private EControlCols body;

	private ETable tbl;

	private ETextArea txtResult;

	private ELayer top;

	private EProgress progress;

	private EButton btn开始;

	private ETextArea txtIDs;

	private ELabelArea lbNote;

	private ELoader lblTotal;

	private ECheck chk编号;

	private ESelect sel产品;

	private ECombo PLogin;

	private ESelectDateRange PDate;

	private ELocalSearch PKind;

	private EButton btn重置;

	private EText txt净收益价;

	private EButton btn改净收益价;

	private ESelect selCur;

	private EText txt最低价;

	public static 任务Meli净收益估算 Instance => Global.Create<任务Meli净收益估算>();

	public 任务Meli净收益估算()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		Dock = DockStyle.Fill;
		tbl.Padding = new Padding(6, 5, 6, 5);
		tbl.Columns.AddRange(new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A55EDA4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6EEB56 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535), 55, 55, ETCellStyle.Image), new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E5B718 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x602E9D9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2), 100, 0, ETheme.Verdana), new ETColumn(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224B010E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A4E848 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010), 80, int.MaxValue, ETheme.Verdana), new ETColumn(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F1C329 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C154484 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1), 80, 0, ETheme.Verdana), new ETColumn(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F217006 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AC2A28 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611), 80, 0, ETheme.Verdana), new ETColumn(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5011E216 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A54654A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39), 70, 150));
		tbl.RowChanged = tbl_RowChanged;
		tbl.RowSetting = tbl_RowSetting;
		sel产品.Items.AddRange(new EItem(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6397464E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2), Color.DarkBlue), new EItem(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D6AA5E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), Color.DarkGreen));
		sel产品.SelectedItem = sel产品.Items[0];
		PLogin.Items = Login.Items;
		if (Member.PID == 7613)
		{
			if (Login.Items.Count > 0)
			{
				PLogin.Visible = true;
				PLogin.NeedText = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F47FD34 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835);
				PLogin.Value = PLogin.Items.FirstKey;
			}
		}
		else if (!Member.只看自己产品)
		{
			PLogin.Visible = Login.Items.Count > 0;
		}
		else
		{
			PLogin.Value = Member.LoginID.ToString();
		}
		PDate.Start.PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DCF933E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5);
		PDate.End.PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BBFB29 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef);
		btn开始.Click += Btn开始_Click;
		btn重置.Click += Btn清理_Click;
		selCur.Enabled = Member.Limit.产品.编辑产品 && Member.Limit.产品.批量替换;
		txt净收益价.Enabled = selCur.Enabled;
		if (!txt净收益价.Enabled)
		{
			return;
		}
		selCur.Items = Curs.Items;
		selCur.SelectedChanged += delegate
		{
			if (selCur.Value.IsNullOrEmpty())
			{
				txt净收益价.PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7087EB0A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834);
			}
			else
			{
				txt净收益价.PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E433BEA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee) + selCur.Text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B383CE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
			}
		};
		selCur.Value = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A55B1BC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39);
		selCur.OnSelectedChanged();
		btn改净收益价.Click += Btn改净收益价_Click;
	}

	private async void Btn清理_Click(object sender, EventArgs e)
	{
		if (EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1705051699 ^ -1939842052 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8)) && EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FBC0DED ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4)))
		{
			ZClient zc = ZClient.Create(ZClient.美客多);
			zc.Loader = lblTotal;
			zc.Button = btn重置;
			zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x501BDDAE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e);
			Para para = GetPara();
			await zc.PostAsync(para);
			zc.Throw();
			DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(14, 1);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A5E1580 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39));
			defaultInterpolatedStringHandler.AppendFormatted(zc.Result);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2633AA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3));
			EDialog.Show(defaultInterpolatedStringHandler.ToStringAndClear(), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-813636989) ^ 0x12356013 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2));
		}
	}

	private async void Btn改净收益价_Click(object sender, EventArgs e)
	{
		Para para = GetPara();
		if (para.low <= 0m)
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D3E0F52 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b));
		}
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(34, 2);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63FAECFB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1));
		defaultInterpolatedStringHandler.AppendFormatted(sel产品.Text);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x764D7875 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835));
		defaultInterpolatedStringHandler.AppendFormatted(para.low);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x639C3E42 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2));
		if (EDialog.Confirm(defaultInterpolatedStringHandler.ToStringAndClear()))
		{
			ZClient zc = ZClient.Create(ZClient.美客多);
			zc.Loader = lblTotal;
			zc.Button = btn改净收益价;
			zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BBB1F7F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018);
			await zc.PostAsync(para);
			zc.Throw();
			defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(14, 1);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78EF33CD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c));
			defaultInterpolatedStringHandler.AppendFormatted(zc.Result);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D649076 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9));
			EDialog.Show(defaultInterpolatedStringHandler.ToStringAndClear(), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A92A4E2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380));
		}
	}

	private void tbl_RowSetting(ETRow tr, Record re)
	{
	}

	private void tbl_RowChanged(ETRow tr, ETRowEventArgs e)
	{
		if (tr != null)
		{
			ETCell eTCell = tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BDD1FB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef)];
			txtResult.ForeColor = eTCell.Color;
			txtResult.Text = eTCell.Text;
			if (tr.Tag is SaleInfo saleInfo)
			{
				ETextArea eTextArea = txtResult;
				DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(7, 2);
				defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xAC94E6E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057));
				defaultInterpolatedStringHandler.AppendFormatted(saleInfo.id);
				defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76478CEB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835));
				defaultInterpolatedStringHandler.AppendFormatted(saleInfo.DetailMessage.IsNullOrEmpty() ? saleInfo.Message : saleInfo.DetailMessage);
				eTextArea.Text = defaultInterpolatedStringHandler.ToStringAndClear();
			}
		}
	}

	private Para GetPara()
	{
		return new Para
		{
			start = PDate.Start.Value,
			end = PDate.End.Value,
			loginid = Valid.ToInt(PLogin.Value),
			ids = txtIDs.Longs,
			localid = PKind.Id,
			from = sel产品.Value.ToInt(),
			cur = selCur.Value,
			low = txt净收益价.Money
		};
	}

	private async Task<List<SaleInfo>> LoadSales()
	{
		tbl.Rows.Clear();
		ZClient zc = ZClient.Create(ZClient.美客多);
		zc.Loader = lblTotal;
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A5E15B6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39);
		Para para = GetPara();
		await zc.PostAsync(para);
		zc.Throw();
		return zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1164AF07 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca)].ToList<SaleInfo>();
	}

	private async void Btn开始_Click(object sender, EventArgs e)
	{
		if (btn开始.Text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2AD4F2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3))
		{
			_busing = true;
			btn开始.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60F11515 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5);
			btn开始.ForeColor = Color.DarkRed;
		}
		else
		{
			_busing = false;
			lblTotal.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-2066591320) ^ 0x45165732 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		}
		try
		{
			if (_busing)
			{
				await Working(await LoadSales());
			}
		}
		finally
		{
			_busing = false;
			btn开始.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC2F7D6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332);
			btn开始.ForeColor = Color.Empty;
		}
	}

	private async Task Working(List<SaleInfo> sales)
	{
		ELoader eLoader = lblTotal;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(11, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB09023 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018));
		defaultInterpolatedStringHandler.AppendFormatted(sales.Count);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x66E8B8C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8));
		eLoader.Text = defaultInterpolatedStringHandler.ToStringAndClear();
		if (sales.Count == 0)
		{
			progress.Visible = false;
			return;
		}
		progress.Visible = true;
		progress.MaxValue = sales.Count;
		int totalPage = (int)Math.Ceiling((double)sales.Count / 10.0);
		for (int page = 0; page < totalPage; page++)
		{
			if (!_busing)
			{
				break;
			}
			int idx = page * 10;
			IEnumerable<SaleInfo> list = sales.Skip(idx).Take(10);
			ELoader eLoader2 = lblTotal;
			defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(1, 2);
			defaultInterpolatedStringHandler.AppendFormatted(idx + 1);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -1139768041 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1));
			defaultInterpolatedStringHandler.AppendFormatted(sales.Count);
			eLoader2.Text = defaultInterpolatedStringHandler.ToStringAndClear();
			progress.Value = idx + 1;
			SetRows(list.ToArray());
			await NetProceedEstimate(list);
			if (_busing)
			{
				ELoader eLoader3 = lblTotal;
				defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(1, 2);
				defaultInterpolatedStringHandler.AppendFormatted(idx + list.Count());
				defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1051CF8E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf));
				defaultInterpolatedStringHandler.AppendFormatted(sales.Count);
				eLoader3.Text = defaultInterpolatedStringHandler.ToStringAndClear();
				progress.Value = idx + 10;
				continue;
			}
			break;
		}
	}

	private async Task NetProceedEstimate(IEnumerable<SaleInfo> sales)
	{
		foreach (SaleInfo sale in sales)
		{
			try
			{
				MeliUrlInfo meliUrlInfo = MeliUrlInfo.Parse(sale.source);
				StringBuilder log = new StringBuilder();
				decimal num = txt最低价.Money;
				if (num <= 0m)
				{
					num = 0m;
				}
				decimal num2 = await MeliNetProceeds.估算净收益(sale.cost, sale.weight, sale.size, meliUrlInfo.Site, num, log);
				if (num2 > 0m)
				{
					sale.netproceed = num2;
				}
				sale.Successed = true;
				sale.Message = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D649738 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9);
				sale.DetailMessage = log.ToString();
				await Task.Delay(10);
			}
			catch (Exception ex)
			{
				sale.Successed = false;
				sale.DetailMessage = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22414352 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2) + ex.Message;
			}
			finally
			{
				SetRows(sale);
			}
		}
		Record record = Record.From(sales.ToList());
		ZClient zc = ZClient.Create(ZClient.美客多);
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x322F63DA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		await zc.PostAsync(record);
		zc.Throw();
	}

	private void SetRows(params SaleInfo[] list)
	{
		foreach (SaleInfo saleInfo in list)
		{
			ETRow eTRow = tbl.Rows.Find(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E30DEAC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6), saleInfo.id.ToString());
			if (eTRow == null)
			{
				saleInfo.Message = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB7F805 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018);
				eTRow = tbl.Rows.Add();
				eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC53638 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332)] = saleInfo.id.ToString();
				eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0F43B5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd)] = saleInfo.thumb;
				eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11652BF9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca)] = saleInfo.source;
				eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -1921083625 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1)] = saleInfo.cost.ToString();
				eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B29562 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53)] = saleInfo.weight.ToString();
				tbl.CurrentRow = eTRow;
			}
			ETCell eTCell = eTRow.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC428FA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332)];
			eTCell.Text = saleInfo.Message;
			eTCell.Color = Color.Empty;
			if (saleInfo.Successed.HasValue)
			{
				eTCell.Color = ((saleInfo.Successed == true) ? Color.Green : Color.Red);
			}
			eTRow.Tag = saleInfo;
		}
		tbl.Refresh();
		tbl.StartLoad();
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
		this.body = new ZYing.UI.EControlCols();
		this.tbl = new ZYing.UI.ETable();
		this.txtResult = new ZYing.UI.ETextArea();
		this.top = new ZYing.UI.ELayer();
		this.lbNote = new ZYing.UI.ELabelArea();
		this.sel产品 = new ZYing.UI.ESelect();
		this.PLogin = new ZYing.UI.ECombo();
		this.PDate = new ZYing.UI.ESelectDateRange();
		this.selCur = new ZYing.UI.ESelect();
		this.txtIDs = new ZYing.UI.ETextArea();
		this.btn重置 = new ZYing.UI.EButton();
		this.btn开始 = new ZYing.UI.EButton();
		this.lblTotal = new ZYing.UI.ELoader();
		this.progress = new ZYing.UI.EProgress();
		this.txt净收益价 = new ZYing.UI.EText();
		this.btn改净收益价 = new ZYing.UI.EButton();
		this.chk编号 = new ZYing.UI.ECheck();
		this.txt最低价 = new ZYing.UI.EText();
		this.PKind = new ZYing.UI.ELocalSearch();
		this.body.SuspendLayout();
		this.top.SuspendLayout();
		base.SuspendLayout();
		this.body.Controls.Add(this.tbl);
		this.body.Controls.Add(this.txtResult);
		this.body.Dock = System.Windows.Forms.DockStyle.Fill;
		this.body.Location = new System.Drawing.Point(15, 117);
		this.body.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.body.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B3CF43 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		this.body.Size = new System.Drawing.Size(1189, 618);
		this.body.TabIndex = 2;
		this.body.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F22116A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a);
		this.tbl.IntegralHeight = false;
		this.tbl.LineStyle = System.Drawing.Drawing2D.DashStyle.Dot;
		this.tbl.Location = new System.Drawing.Point(0, 0);
		this.tbl.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.tbl.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E304C8A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.tbl.ShowBorder = true;
		this.tbl.Size = new System.Drawing.Size(589, 617);
		this.tbl.TabIndex = 2;
		this.txtResult.FlowAlign = ZYing.UI.FlowAlignment.Fill;
		this.txtResult.ForeColor = System.Drawing.Color.Red;
		this.txtResult.Location = new System.Drawing.Point(599, 0);
		this.txtResult.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.txtResult.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A9272BC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380);
		this.txtResult.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtResult.Size = new System.Drawing.Size(589, 617);
		this.txtResult.TabIndex = 3;
		this.txtResult.TabStop = false;
		this.top.Controls.Add(this.lbNote);
		this.top.Controls.Add(this.sel产品);
		this.top.Controls.Add(this.PLogin);
		this.top.Controls.Add(this.PDate);
		this.top.Controls.Add(this.PKind);
		this.top.Controls.Add(this.selCur);
		this.top.Controls.Add(this.txtIDs);
		this.top.Controls.Add(this.btn重置);
		this.top.Controls.Add(this.btn开始);
		this.top.Controls.Add(this.txt最低价);
		this.top.Controls.Add(this.lblTotal);
		this.top.Controls.Add(this.progress);
		this.top.Controls.Add(this.txt净收益价);
		this.top.Controls.Add(this.btn改净收益价);
		this.top.Dock = System.Windows.Forms.DockStyle.Top;
		this.top.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.top.Location = new System.Drawing.Point(15, 12);
		this.top.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D795F6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda);
		this.top.Padding = new System.Windows.Forms.Padding(0, 0, 0, 6);
		this.top.Size = new System.Drawing.Size(1189, 105);
		this.top.TabIndex = 2;
		this.lbNote.AutoSize = false;
		this.lbNote.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.lbNote.ForeColor = System.Drawing.Color.MediumBlue;
		this.lbNote.LineSpacing = 1;
		this.lbNote.Location = new System.Drawing.Point(3, 3);
		this.lbNote.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.lbNote.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B2FC97 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e);
		this.lbNote.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.lbNote.Size = new System.Drawing.Size(1182, 28);
		this.lbNote.TabIndex = 363;
		this.lbNote.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643DCA0A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		this.sel产品.AutoHeight = true;
		this.sel产品.FlowBreak = false;
		this.sel产品.FlowWrap = false;
		this.sel产品.Location = new System.Drawing.Point(0, 34);
		this.sel产品.Margin = new System.Windows.Forms.Padding(0, 3, 3, 3);
		this.sel产品.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C75D36 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.sel产品.NeedText = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x502010B9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841);
		this.sel产品.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.sel产品.Size = new System.Drawing.Size(134, 28);
		this.sel产品.TabIndex = 388;
		this.PLogin.AutoHeight = true;
		this.PLogin.FlowBreak = false;
		this.PLogin.IsExact = true;
		this.PLogin.Location = new System.Drawing.Point(140, 34);
		this.PLogin.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB75097 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4);
		this.PLogin.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PLogin.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527FB97D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1);
		this.PLogin.Size = new System.Drawing.Size(142, 28);
		this.PLogin.TabIndex = 387;
		this.PLogin.Visible = false;
		this.PDate.FlowBreak = false;
		this.PDate.Location = new System.Drawing.Point(288, 34);
		this.PDate.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A92C066 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380);
		this.PDate.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PDate.Size = new System.Drawing.Size(222, 28);
		this.PDate.TabIndex = 385;
		this.selCur.Always = true;
		this.selCur.AutoHeight = true;
		this.selCur.DefaultValue = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E305B10 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.selCur.FlowBreak = false;
		this.selCur.Location = new System.Drawing.Point(830, 34);
		this.selCur.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6FB072 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.selCur.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.selCur.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224DA372 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2);
		this.selCur.Size = new System.Drawing.Size(81, 28);
		this.selCur.TabIndex = 141;
		this.txtIDs.AcceptReturn = false;
		this.txtIDs.Anchor = System.Windows.Forms.AnchorStyles.None;
		this.txtIDs.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB73D11 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4);
		this.txtIDs.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.txtIDs.FlowBreak = false;
		this.txtIDs.Indent = 1;
		this.txtIDs.Location = new System.Drawing.Point(917, 34);
		this.txtIDs.MaximumSize = new System.Drawing.Size(10000, 27);
		this.txtIDs.MaxLength = 10000;
		this.txtIDs.MaxLines = 1;
		this.txtIDs.MinimumSize = new System.Drawing.Size(0, 28);
		this.txtIDs.MinLines = 1;
		this.txtIDs.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA7AB75 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
		this.txtIDs.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtIDs.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x7DEA9D31 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.txtIDs.PostVisible = true;
		this.txtIDs.Size = new System.Drawing.Size(160, 28);
		this.txtIDs.TabIndex = 519;
		this.txtIDs.TabStop = false;
		this.txtIDs.VisibleDesign = false;
		this.btn重置.ForeColor = System.Drawing.Color.DarkRed;
		this.btn重置.Location = new System.Drawing.Point(1083, 34);
		this.btn重置.Margin = new System.Windows.Forms.Padding(3, 3, 5, 3);
		this.btn重置.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63609075 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3);
		this.btn重置.Padding = new System.Windows.Forms.Padding(8, 4, 8, 4);
		this.btn重置.Size = new System.Drawing.Size(100, 27);
		this.btn重置.TabIndex = 4;
		this.btn重置.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FDF321 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5);
		this.btn开始.FlowBreak = false;
		this.btn开始.Location = new System.Drawing.Point(3, 68);
		this.btn开始.Margin = new System.Windows.Forms.Padding(3, 3, 5, 3);
		this.btn开始.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A55DA48 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39);
		this.btn开始.Padding = new System.Windows.Forms.Padding(8, 4, 8, 4);
		this.btn开始.Size = new System.Drawing.Size(72, 27);
		this.btn开始.TabIndex = 3;
		this.btn开始.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643DCABC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		this.lblTotal.FlowBreak = false;
		this.lblTotal.ForeColor = System.Drawing.Color.DarkBlue;
		this.lblTotal.Location = new System.Drawing.Point(303, 68);
		this.lblTotal.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F16519 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1);
		this.lblTotal.Size = new System.Drawing.Size(174, 28);
		this.lblTotal.TabIndex = 5;
		this.progress.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.progress.FlowBreak = false;
		this.progress.Location = new System.Drawing.Point(480, 71);
		this.progress.Margin = new System.Windows.Forms.Padding(0, 6, 0, 4);
		this.progress.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A8C849C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21);
		this.progress.Size = new System.Drawing.Size(471, 17);
		this.progress.TabIndex = 3;
		this.progress.TextAlign = System.Drawing.StringAlignment.Near;
		this.progress.Visible = false;
		this.txt净收益价.Anchor = System.Windows.Forms.AnchorStyles.None;
		this.txt净收益价.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACE994C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057);
		this.txt净收益价.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.txt净收益价.FlowBreak = false;
		this.txt净收益价.Indent = 1;
		this.txt净收益价.Location = new System.Drawing.Point(954, 68);
		this.txt净收益价.MaximumSize = new System.Drawing.Size(10000, 27);
		this.txt净收益价.MaxLength = 10000;
		this.txt净收益价.MinimumSize = new System.Drawing.Size(0, 28);
		this.txt净收益价.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x322019CC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		this.txt净收益价.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txt净收益价.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76473D2D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
		this.txt净收益价.PostVisible = true;
		this.txt净收益价.Size = new System.Drawing.Size(123, 28);
		this.txt净收益价.TabIndex = 521;
		this.txt净收益价.TabStop = false;
		this.txt净收益价.ValidType = ZYing.UI.EValidType.Money;
		this.txt净收益价.VisibleDesign = false;
		this.btn改净收益价.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.btn改净收益价.ForeColor = System.Drawing.Color.Blue;
		this.btn改净收益价.Location = new System.Drawing.Point(1083, 68);
		this.btn改净收益价.Margin = new System.Windows.Forms.Padding(3, 3, 5, 3);
		this.btn改净收益价.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B41735 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		this.btn改净收益价.Padding = new System.Windows.Forms.Padding(8, 4, 8, 4);
		this.btn改净收益价.Size = new System.Drawing.Size(100, 27);
		this.btn改净收益价.TabIndex = 520;
		this.btn改净收益价.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -560002784 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd);
		this.chk编号.Indent = 76;
		this.chk编号.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468F0BDC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6);
		this.chk编号.Location = new System.Drawing.Point(3, 3);
		this.chk编号.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-713251553) ^ 0x7695F0B7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.chk编号.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk编号.Size = new System.Drawing.Size(419, 28);
		this.chk编号.TabIndex = 520;
		this.chk编号.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E37EA1A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.txt最低价.Bold = true;
		this.txt最低价.Cur = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA58BF3F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1);
		this.txt最低价.FlowBreak = false;
		this.txt最低价.Fonts = ZYing.UI.EFonts.VerdanaBold;
		this.txt最低价.Location = new System.Drawing.Point(83, 68);
		this.txt最低价.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D12C5C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda);
		this.txt最低价.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txt最低价.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -1633126866 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
		this.txt最低价.RightColor = System.Drawing.Color.DarkGreen;
		this.txt最低价.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6626B96 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8);
		this.txt最低价.Size = new System.Drawing.Size(214, 28);
		this.txt最低价.TabIndex = 3;
		this.txt最低价.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F60AB1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1);
		this.txt最低价.ValidType = ZYing.UI.EValidType.Money;
		this.PKind.AutoHeight = true;
		this.PKind.FlowBreak = false;
		this.PKind.Location = new System.Drawing.Point(516, 34);
		this.PKind.MinimumSize = new System.Drawing.Size(222, 0);
		this.PKind.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x50275C5B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841);
		this.PKind.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PKind.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B0BF79 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e);
		this.PKind.Size = new System.Drawing.Size(308, 28);
		this.PKind.TabIndex = 386;
		base.Controls.Add(this.body);
		base.Controls.Add(this.top);
		this.EIcos = ZYing.IcoEnums.NULL;
		this.FlowAlign = ZYing.UI.FlowAlignment.Left;
		this.Fonts = ZYing.UI.EFonts.微软雅黑;
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D687682 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9);
		this.Padding = new System.Windows.Forms.Padding(15, 12, 15, 12);
		base.Size = new System.Drawing.Size(1220, 748);
		this.body.ResumeLayout(false);
		this.top.ResumeLayout(false);
		this.top.PerformLayout();
		base.ResumeLayout(false);
	}
}
