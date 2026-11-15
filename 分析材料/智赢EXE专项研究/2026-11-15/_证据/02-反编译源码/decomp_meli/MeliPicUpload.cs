using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Drawing;
using System.Drawing.Drawing2D;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Threading.Tasks;
using System.Windows.Forms;
using Enumer;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Data;
using ZYing.Meli;
using ZYing.Meli.Model;
using ZYing.UI;

namespace ZYing.Channel;

[ToolboxItem(false)]
public class Meli图片上传 : ELayer
{
	private class SaleInfo
	{
		private List<MeliPicInfo> _meliPics;

		public int Id { get; set; }

		public int Accid { get; set; }

		public string[] Pic { get; set; }

		public bool? Successed { get; set; }

		public string Message { get; set; }

		public string DetailMessage { get; set; }

		public List<MeliPicInfo> MeliPics
		{
			get
			{
				if (_meliPics != null)
				{
					return _meliPics;
				}
				_meliPics = Pic.Select((string x) => new MeliPicInfo
				{
					Saleid = Id,
					Accid = Accid,
					Url = x
				}).ToList();
				return _meliPics;
			}
		}

		public Record MeliPicRecord => MeliPics?.ToRecord() ?? new Record();

		public SaleInfo()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}
	}

	private class MeliPicInfo
	{
		public string Url { get; set; }

		public int Saleid { get; set; }

		public int Accid { get; set; }

		public string Picid { get; set; }

		public string Meliurl { get; set; }

		public int Width { get; set; }

		public int Height { get; set; }

		public int Stat { get; set; }

		public string Msg { get; set; }

		public MeliPicInfo()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			Stat = 1;
			base..ctor();
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass12_0
	{
		public Meli图片上传 <>4__this;

		public int aid;

		public <>c__DisplayClass12_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass14_0
	{
		public Meli图片上传 <>4__this;

		public MeliApi api;

		public <>c__DisplayClass14_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}

		internal async Task <UploadImg>b__0(IEnumerable<MeliPicInfo> list, int num)
		{
			Queue<MeliPicInfo> picQueue = (Queue<MeliPicInfo>)list;
			while (picQueue.Count > 0)
			{
				MeliPicInfo pic = picQueue.Dequeue();
				await <>4__this.UploadImgCore(api, pic);
			}
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass15_0
	{
		public MeliPicInfo pic;

		public <>c__DisplayClass15_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}

		internal void <UploadImgCore>b__0(Exception ex)
		{
			pic.Msg = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AA4BB4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611) + ex.Message;
			pic.Stat = 9;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass17_0
	{
		public Record re;

		public Func<ETRow, bool> <>9__0;

		public <>c__DisplayClass17_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}

		internal bool <BtnSubmit_Click>b__0(ETRow x)
		{
			return x[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5010FCD4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e)] == re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB79245 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4));
		}
	}

	public const string Key = "meli_pic_upload";

	public const string Caption = "图片上传";

	private bool _busing;

	private IContainer components;

	private EControlCols body;

	private ETable tbl;

	private ELayer top;

	private EProgress progress;

	private EButton btn开始;

	private ETextArea txtIDs;

	private ELabelArea lbNote;

	private ELoader lblTotal;

	private ECheck chk编号;

	private ECombo PLogin;

	private ESelectDateRange PDate;

	private ELocalSearch PKind;

	private ECombo PSource;

	private EButton btnSubmit;

	private ELayer buttomLy;

	private ELoader eLoader1;

	private ELayer eLayer1;

	private ELayer eLayer2;

	private ETable tblMeliPics;

	public static Meli图片上传 Instance => Global.Create<Meli图片上传>();

	public Meli图片上传()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		tbl.RowHeight = 45;
		tbl.Columns.AddRange(new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA68297 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E54465 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c), 0, 0, hidden: true), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5111166F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BC1A2D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef), 55, 150, ETCellStyle.Images)
		{
			IsDrop = true
		}, new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB02293 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10346BB2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d), 60, 80, ETheme.Verdana), new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A32948 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -263105687 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf), 100, 0, ETheme.VerdanaBold), new ETColumn(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -2111064407 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -1453236623 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689), 80, 0, ETheme.微软雅黑));
		tbl.RowChanged = tbl_RowChanged;
		tbl.RowSetting = tbl_RowSetting;
		tbl.RowPicked = delegate
		{
		};
		tblMeliPics.RowHeight = 55;
		tblMeliPics.Columns.AddRange(new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63674143 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76464EB1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835), 0, 0, hidden: true), new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16358805 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D6E0F2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), 0, 0, hidden: true), new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527F153B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7086454A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834), 0, 0, hidden: true), new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x501031A8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x66566C4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8), 0, 0, hidden: true), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11646E49 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BC1A2D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef), 50, 60, ETCellStyle.Image)
		{
			IsDrop = true
		}, new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AA483A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x50171FC6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e), 100, 300, ETCellStyle.Multi), new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA030BB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA9F8AC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b), 100, int.MaxValue, ETCellStyle.Multi), new ETColumn(10, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x105143D0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x32266B06 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594), 45, 45, ETCellStyle.OP, ETheme.微软雅黑));
		tblMeliPics.RowCreating = tblPro_RowCreating;
		tblMeliPics.RowSetting = tblPro_RowSetting;
		tblMeliPics.CellOPClick = tblPro_CellOPClick;
		PSource.Items = Enumer.Web.Items;
		PLogin.Items = Login.Items;
		if (!Member.只看自己产品)
		{
			PLogin.Visible = Login.Items.Count > 0;
		}
		else
		{
			PLogin.Value = Member.LoginID.ToString();
		}
		PDate.Start.PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A5201E0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39);
		PDate.End.PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE31A96 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde);
		btn开始.Click += Btn开始_Click;
		btnSubmit.Click += BtnSubmit_Click;
	}

	private void tbl_RowSetting(ETRow tr, Record re)
	{
		string[] array = re.Gets(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA9F8CE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b));
		tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FB4D85 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5)] = array.Take(3).Join(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC5BB5A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332));
		string name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CDB09F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(4, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0F51EB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd));
		defaultInterpolatedStringHandler.AppendFormatted(array.Length);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A4CA60 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010));
		tr[name] = defaultInterpolatedStringHandler.ToStringAndClear();
	}

	private void tbl_RowChanged(ETRow tr, ETRowEventArgs e)
	{
		if (tr != null)
		{
			_ = tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA592441 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1)];
			if (tr.Tag is SaleInfo saleInfo)
			{
				tblMeliPics.DataSource = saleInfo.MeliPicRecord;
			}
		}
	}

	private void tblPro_RowCreating(ETRow tr)
	{
		tr.LastCell.AddRange(new UIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB45905 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018), IcoEnums.go));
	}

	private void tblPro_RowSetting(ETRow tr, Record re)
	{
		int num = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224AAF8C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2));
		int num2 = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3A3B7E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273));
		string value = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E25438 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce));
		int num3 = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2DEBFE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3));
		ETCellMulti eTCellMulti = tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A542AC2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39)] as ETCellMulti;
		eTCellMulti.Color = Color.Empty;
		switch (num3)
		{
		case 1:
			eTCellMulti.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BB2C67 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef);
			break;
		case 2:
			eTCellMulti.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AA4B86 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611);
			eTCellMulti.Color = Color.DarkBlue;
			break;
		case 10:
			eTCellMulti.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75C881A1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
			eTCellMulti.Color = Color.DarkGreen;
			break;
		case 9:
			eTCellMulti.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F6299D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1) + re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -1035286559 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535));
			eTCellMulti.Color = Color.Red;
			break;
		}
		string subText;
		if (num <= 0 || num2 <= 0)
		{
			subText = "";
		}
		else
		{
			DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(3, 2);
			defaultInterpolatedStringHandler.AppendFormatted(num);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA9D0D4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b));
			defaultInterpolatedStringHandler.AppendFormatted(num2);
			subText = defaultInterpolatedStringHandler.ToStringAndClear();
		}
		eTCellMulti.SubText = subText;
		((ETCellOP)tr.LastCell).SetVisible(!value.IsNullOrEmpty(), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63937852 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2));
	}

	private async void tblPro_CellOPClick(ETCellOP c, UIcon ico)
	{
		if (ico.Name == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E0757B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c))
		{
			string text = c.Row[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-2066591320) ^ 0x6F9A5639 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef)];
			if (!text.IsNullOrEmpty())
			{
				Tool.Start(text);
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
			kind = PSource.Value.ToInt()
		};
	}

	private async void Btn开始_Click(object sender, EventArgs e)
	{
		<>c__DisplayClass12_0 CS$<>8__locals6 = new <>c__DisplayClass12_0();
		CS$<>8__locals6.<>4__this = this;
		if (btn开始.Text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E43ACBD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689))
		{
			_busing = true;
			btn开始.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86DDC53 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
			btn开始.ForeColor = Color.DarkRed;
			tbl.Rows.Clear();
			tblMeliPics.Rows.Clear();
		}
		else
		{
			_busing = false;
			lblTotal.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5274ECAB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1);
		}
		try
		{
			if (_busing)
			{
				CS$<>8__locals6.aid = 0;
				await Working(await LoadSales(), CS$<>8__locals6.aid);
			}
		}
		finally
		{
			_busing = false;
			btn开始.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51150B09 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
			btn开始.ForeColor = Color.Empty;
		}
		async Task<List<SaleInfo>> LoadSales()
		{
			ZClient zc = ZClient.Create(ZClient.美客多);
			zc.Loader = CS$<>8__locals6.<>4__this.lblTotal;
			zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x40801A0A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44);
			Para para = CS$<>8__locals6.<>4__this.GetPara();
			await zc.PostAsync(para);
			zc.Throw();
			Record record = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224A85F0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2)];
			CS$<>8__locals6.aid = (int)zc.Result;
			return record.ToList<SaleInfo>();
		}
	}

	private async Task Working(List<SaleInfo> sales, int aid)
	{
		ELoader eLoader = lblTotal;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(12, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B7C4EC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53));
		defaultInterpolatedStringHandler.AppendFormatted(sales.Count);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22415592 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2));
		eLoader.Text = defaultInterpolatedStringHandler.ToStringAndClear();
		if (sales.Count == 0)
		{
			progress.Visible = false;
			return;
		}
		progress.Visible = true;
		progress.MaxValue = sales.Count;
		progress.Value = 0;
		for (int idx = 0; idx < sales.Count; idx++)
		{
			if (!_busing)
			{
				break;
			}
			SaleInfo sale = sales[idx];
			sale.Accid = aid;
			progress.Value = idx + 1;
			ELoader eLoader2 = lblTotal;
			defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(1, 2);
			defaultInterpolatedStringHandler.AppendFormatted(idx + 1);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6E1CDE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535));
			defaultInterpolatedStringHandler.AppendFormatted(sales.Count);
			eLoader2.Text = defaultInterpolatedStringHandler.ToStringAndClear();
			SetRow(sale);
			await UploadImg(sale);
			SetRow(sale);
		}
	}

	private async Task UploadImg(SaleInfo sale)
	{
		<>c__DisplayClass14_0 CS$<>8__locals4 = new <>c__DisplayClass14_0();
		CS$<>8__locals4.<>4__this = this;
		CS$<>8__locals4.api = await MeliApiBuilder.Create(sale.Accid);
		await new Queue<MeliPicInfo>(sale.MeliPics).ParallelAsync(async (IEnumerable<MeliPicInfo> list, int num) =>
		{
			Queue<MeliPicInfo> picQueue = (Queue<MeliPicInfo>)list;
			while (picQueue.Count > 0)
			{
				MeliPicInfo pic = picQueue.Dequeue();
				await CS$<>8__locals4.<>4__this.UploadImgCore(CS$<>8__locals4.api, pic);
			}
		}, 5);
		Record re = sale.MeliPicRecord;
		ZClient zc = ZClient.Create(ZClient.美客多);
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63FAFE4D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1);
		await zc.PostAsync(re);
		if (zc.IsErr)
		{
			sale.Successed = false;
			sale.Message = zc.Message;
			return;
		}
		sale.Successed = true;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(7, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -776513561 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53));
		defaultInterpolatedStringHandler.AppendFormatted(re.Count);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F2294B0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a));
		sale.Message = defaultInterpolatedStringHandler.ToStringAndClear();
	}

	private async Task UploadImgCore(MeliApi api, MeliPicInfo pic)
	{
		<>c__DisplayClass15_0 CS$<>8__locals21 = new <>c__DisplayClass15_0();
		CS$<>8__locals21.pic = pic;
		CS$<>8__locals21.pic.Stat = 2;
		CS$<>8__locals21.pic.Msg = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B8E3C7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		if (CS$<>8__locals21.pic.Url.IndexOf(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45DDDB78 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), StringComparison.OrdinalIgnoreCase) > -1)
		{
			CS$<>8__locals21.pic.Msg = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86DDFE5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
			CS$<>8__locals21.pic.Stat = 10;
			CS$<>8__locals21.pic.Picid = Enumer.Meli.GetImageId(CS$<>8__locals21.pic.Url);
			CS$<>8__locals21.pic.Meliurl = CS$<>8__locals21.pic.Url;
			return;
		}
		try
		{
			MeliPicture meliPicture = await api.UploadPic(CS$<>8__locals21.pic.Url, (Exception ex2) =>
			{
				CS$<>8__locals21.pic.Msg = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AA4BB4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611) + ex2.Message;
				CS$<>8__locals21.pic.Stat = 9;
			});
			if (meliPicture != null)
			{
				CS$<>8__locals21.pic.Msg = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x764374D9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
				CS$<>8__locals21.pic.Stat = 10;
				CS$<>8__locals21.pic.Picid = meliPicture.id;
				CS$<>8__locals21.pic.Meliurl = meliPicture.SecureUrl;
				CS$<>8__locals21.pic.Height = meliPicture.Height;
				CS$<>8__locals21.pic.Width = meliPicture.Width;
			}
		}
		catch (Exception ex)
		{
			CS$<>8__locals21.pic.Msg = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA5FD9F9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1) + ex.Message;
			CS$<>8__locals21.pic.Stat = 9;
		}
	}

	private ETRow SetRow(SaleInfo item)
	{
		ETRow eTRow = tbl.Rows.Find(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643A0B52 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848), item.Id.ToString());
		if (eTRow == null)
		{
			item.Message = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA030D7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
			eTRow = tbl.Rows.Add();
			eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x665ABB8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8)] = item.Id.ToString();
			eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E318F84 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6)] = item.Pic.Join(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B35A4D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552));
			ETRow eTRow2 = eTRow;
			string name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C155896 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
			string[] pic = item.Pic;
			object value;
			if (pic == null)
			{
				value = null;
			}
			else
			{
				IEnumerable<string> enumerable = pic.Take(3);
				value = ((enumerable != null) ? enumerable.Join(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63979506 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2)) : null);
			}
			eTRow2[name] = (string)value;
			ETRow eTRow3 = eTRow;
			string name2 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E44E026 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee);
			DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(4, 1);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5826A719 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd));
			defaultInterpolatedStringHandler.AppendFormatted(item.Pic.Length);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E3F665 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c));
			eTRow3[name2] = defaultInterpolatedStringHandler.ToStringAndClear();
			tbl.CurrentRow = eTRow;
		}
		ETCell eTCell = eTRow.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA592441 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1)];
		eTCell.Text = item.Message;
		eTCell.Color = Color.Empty;
		if (item.Successed.HasValue)
		{
			eTCell.Color = ((item.Successed == true) ? Color.Green : Color.Red);
		}
		eTRow.Tag = item;
		tbl.Refresh();
		tbl.StartLoad();
		return eTRow;
	}

	private async void BtnSubmit_Click(object sender, EventArgs e)
	{
		<>c__DisplayClass17_0 CS$<>8__locals9 = new <>c__DisplayClass17_0();
		ETRow[] rows = tbl.RowsPick;
		if (rows == null || rows.Length == 0)
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B54720 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2));
		}
		CS$<>8__locals9.re = new Record();
		ETRow[] array = rows;
		foreach (ETRow obj in array)
		{
			SaleInfo saleInfo = obj.Tag as SaleInfo;
			CS$<>8__locals9.re.AddRow();
			CS$<>8__locals9.re.Columns.AddInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5110D07F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7)).Set(saleInfo.Id);
			string text = obj[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E5769D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c)];
			if (!text.IsNullOrEmpty() && Uri.IsWellFormedUriString(text, UriKind.Absolute))
			{
				CS$<>8__locals9.re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D2F488 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda)).Set(text);
			}
		}
		if (CS$<>8__locals9.re.Count == 0)
		{
			EDialog.Show(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x502DEAAD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DC8B916 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5));
			return;
		}
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(19, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x66F6650 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8));
		defaultInterpolatedStringHandler.AppendFormatted(CS$<>8__locals9.re.Count);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2240AEB8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2));
		if (!EDialog.Confirm(defaultInterpolatedStringHandler.ToStringAndClear()))
		{
			return;
		}
		ZClient zc = ZClient.Create();
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68A764C8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611);
		zc.Loader = eLoader1;
		zc.Button = btnSubmit;
		await zc.PostAsync(CS$<>8__locals9.re);
		zc.Throw();
		while (CS$<>8__locals9.re.Read())
		{
			ETRow eTRow = rows.FirstOrDefault((ETRow x) => x[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5010FCD4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e)] == CS$<>8__locals9.re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB79245 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4)));
			if (eTRow != null)
			{
				tbl.Rows.Remove(eTRow);
			}
		}
		tblMeliPics.Rows.Clear();
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
		this.chk编号 = new ZYing.UI.ECheck();
		this.top = new ZYing.UI.ELayer();
		this.lbNote = new ZYing.UI.ELabelArea();
		this.PLogin = new ZYing.UI.ECombo();
		this.PKind = new ZYing.UI.ELocalSearch();
		this.PSource = new ZYing.UI.ECombo();
		this.PDate = new ZYing.UI.ESelectDateRange();
		this.txtIDs = new ZYing.UI.ETextArea();
		this.btn开始 = new ZYing.UI.EButton();
		this.lblTotal = new ZYing.UI.ELoader();
		this.progress = new ZYing.UI.EProgress();
		this.eLayer1 = new ZYing.UI.ELayer();
		this.body = new ZYing.UI.EControlCols();
		this.tbl = new ZYing.UI.ETable();
		this.eLayer2 = new ZYing.UI.ELayer();
		this.tblMeliPics = new ZYing.UI.ETable();
		this.buttomLy = new ZYing.UI.ELayer();
		this.btnSubmit = new ZYing.UI.EButton();
		this.eLoader1 = new ZYing.UI.ELoader();
		this.top.SuspendLayout();
		this.eLayer1.SuspendLayout();
		this.body.SuspendLayout();
		this.eLayer2.SuspendLayout();
		this.buttomLy.SuspendLayout();
		base.SuspendLayout();
		this.chk编号.Indent = 76;
		this.chk编号.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10352F7A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
		this.chk编号.Location = new System.Drawing.Point(3, 3);
		this.chk编号.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E44431B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		this.chk编号.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk编号.Size = new System.Drawing.Size(419, 28);
		this.chk编号.TabIndex = 520;
		this.chk编号.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CA9F89 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		this.top.Controls.Add(this.lbNote);
		this.top.Controls.Add(this.PLogin);
		this.top.Controls.Add(this.PKind);
		this.top.Controls.Add(this.PSource);
		this.top.Controls.Add(this.PDate);
		this.top.Controls.Add(this.txtIDs);
		this.top.Controls.Add(this.btn开始);
		this.top.Controls.Add(this.lblTotal);
		this.top.Controls.Add(this.progress);
		this.top.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.top.Location = new System.Drawing.Point(18, 15);
		this.top.MaxHeight = 252;
		this.top.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C11308 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.top.Size = new System.Drawing.Size(1183, 118);
		this.top.TabIndex = 2;
		this.lbNote.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.lbNote.ForeColor = System.Drawing.Color.MediumBlue;
		this.lbNote.LineSpacing = 1;
		this.lbNote.Location = new System.Drawing.Point(3, 3);
		this.lbNote.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.lbNote.MaxLines = 2;
		this.lbNote.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C16CAD0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.lbNote.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.lbNote.Size = new System.Drawing.Size(1176, 47);
		this.lbNote.TabIndex = 363;
		this.lbNote.Tag = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CD89A3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		this.lbNote.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51178AAD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
		this.PLogin.AutoHeight = true;
		this.PLogin.FlowBreak = false;
		this.PLogin.IsExact = true;
		this.PLogin.Location = new System.Drawing.Point(3, 53);
		this.PLogin.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -1982556217 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834);
		this.PLogin.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PLogin.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -2090273967 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273);
		this.PLogin.Size = new System.Drawing.Size(127, 28);
		this.PLogin.TabIndex = 387;
		this.PLogin.Visible = false;
		this.PKind.AutoHeight = true;
		this.PKind.FlowBreak = false;
		this.PKind.Location = new System.Drawing.Point(136, 53);
		this.PKind.MinimumSize = new System.Drawing.Size(222, 0);
		this.PKind.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x32243902 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		this.PKind.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PKind.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C14893E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.PKind.Size = new System.Drawing.Size(257, 28);
		this.PKind.TabIndex = 386;
		this.PSource.AutoHeight = true;
		this.PSource.FlowBreak = false;
		this.PSource.IsExact = true;
		this.PSource.Location = new System.Drawing.Point(399, 53);
		this.PSource.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1166C57F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca);
		this.PSource.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PSource.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -1252555381 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		this.PSource.Size = new System.Drawing.Size(83, 28);
		this.PSource.TabIndex = 244;
		this.PDate.FlowBreak = false;
		this.PDate.Location = new System.Drawing.Point(488, 53);
		this.PDate.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A554F4C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39);
		this.PDate.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PDate.Size = new System.Drawing.Size(222, 28);
		this.PDate.TabIndex = 385;
		this.txtIDs.AcceptReturn = false;
		this.txtIDs.Anchor = System.Windows.Forms.AnchorStyles.None;
		this.txtIDs.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F40DED0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835);
		this.txtIDs.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.txtIDs.Indent = 1;
		this.txtIDs.Location = new System.Drawing.Point(716, 53);
		this.txtIDs.MaximumSize = new System.Drawing.Size(10000, 27);
		this.txtIDs.MaxLength = 10000;
		this.txtIDs.MaxLines = 1;
		this.txtIDs.MinimumSize = new System.Drawing.Size(0, 28);
		this.txtIDs.MinLines = 1;
		this.txtIDs.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -1936181184 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
		this.txtIDs.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtIDs.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C131034 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.txtIDs.PostVisible = true;
		this.txtIDs.Size = new System.Drawing.Size(463, 28);
		this.txtIDs.TabIndex = 519;
		this.txtIDs.TabStop = false;
		this.txtIDs.VisibleDesign = false;
		this.btn开始.FlowBreak = false;
		this.btn开始.Location = new System.Drawing.Point(3, 87);
		this.btn开始.Margin = new System.Windows.Forms.Padding(3, 3, 5, 3);
		this.btn开始.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527F3659 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1);
		this.btn开始.Padding = new System.Windows.Forms.Padding(8, 4, 8, 4);
		this.btn开始.Size = new System.Drawing.Size(72, 27);
		this.btn开始.TabIndex = 3;
		this.btn开始.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F4A8DB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1);
		this.lblTotal.FlowBreak = false;
		this.lblTotal.ForeColor = System.Drawing.Color.DarkBlue;
		this.lblTotal.Location = new System.Drawing.Point(83, 87);
		this.lblTotal.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F16519 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1);
		this.lblTotal.Size = new System.Drawing.Size(174, 28);
		this.lblTotal.TabIndex = 5;
		this.progress.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.progress.FlowBreak = false;
		this.progress.Location = new System.Drawing.Point(260, 90);
		this.progress.Margin = new System.Windows.Forms.Padding(0, 6, 0, 4);
		this.progress.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A163FA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010);
		this.progress.Size = new System.Drawing.Size(922, 17);
		this.progress.TabIndex = 3;
		this.progress.TextAlign = System.Drawing.StringAlignment.Near;
		this.progress.Visible = false;
		this.eLayer1.Controls.Add(this.body);
		this.eLayer1.FlowAlign = ZYing.UI.FlowAlignment.Fill;
		this.eLayer1.Location = new System.Drawing.Point(15, 136);
		this.eLayer1.Margin = new System.Windows.Forms.Padding(0);
		this.eLayer1.MaxHeight = 601;
		this.eLayer1.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527FAEDD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1);
		this.eLayer1.Size = new System.Drawing.Size(1189, 601);
		this.eLayer1.TabIndex = 6;
		this.body.Controls.Add(this.tbl);
		this.body.Controls.Add(this.eLayer2);
		this.body.FlowAlign = ZYing.UI.FlowAlignment.Fill;
		this.body.Location = new System.Drawing.Point(3, 0);
		this.body.Margin = new System.Windows.Forms.Padding(3, 0, 3, 0);
		this.body.MaxFix = 1000;
		this.body.MaxHeight = 600;
		this.body.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B29122 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53);
		this.body.Size = new System.Drawing.Size(1182, 600);
		this.body.Space = 2;
		this.body.TabIndex = 2;
		this.body.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x647E9B47 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8);
		this.tbl.FlowAlign = ZYing.UI.FlowAlignment.Fill;
		this.tbl.IntegralHeight = false;
		this.tbl.LineStyle = System.Drawing.Drawing2D.DashStyle.Dot;
		this.tbl.Location = new System.Drawing.Point(0, 0);
		this.tbl.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.tbl.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68D06A96 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6);
		this.tbl.ShowBorder = true;
		this.tbl.Size = new System.Drawing.Size(590, 599);
		this.tbl.TabIndex = 2;
		this.tbl.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C131022 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.eLayer2.Controls.Add(this.tblMeliPics);
		this.eLayer2.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.eLayer2.Location = new System.Drawing.Point(592, 0);
		this.eLayer2.MaxHeight = 326;
		this.eLayer2.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3A4BB4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273);
		this.eLayer2.Size = new System.Drawing.Size(590, 599);
		this.eLayer2.TabIndex = 7;
		this.tblMeliPics.FlowAlign = ZYing.UI.FlowAlignment.Fill;
		this.tblMeliPics.IntegralHeight = false;
		this.tblMeliPics.LineStyle = System.Drawing.Drawing2D.DashStyle.Dot;
		this.tblMeliPics.Location = new System.Drawing.Point(3, 3);
		this.tblMeliPics.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.tblMeliPics.MaxHeight = 595;
		this.tblMeliPics.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D32CA12 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
		this.tblMeliPics.ShowBorder = true;
		this.tblMeliPics.Size = new System.Drawing.Size(583, 595);
		this.tblMeliPics.TabIndex = 7;
		this.tblMeliPics.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-813636989) ^ 0x7BC8C69E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018);
		this.buttomLy.Controls.Add(this.btnSubmit);
		this.buttomLy.Controls.Add(this.eLoader1);
		this.buttomLy.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.buttomLy.Location = new System.Drawing.Point(18, 740);
		this.buttomLy.MaxHeight = 326;
		this.buttomLy.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E323C9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c);
		this.buttomLy.Padding = new System.Windows.Forms.Padding(0, 5, 0, 5);
		this.buttomLy.Size = new System.Drawing.Size(1183, 44);
		this.buttomLy.TabIndex = 5;
		this.btnSubmit.Enabled = false;
		this.btnSubmit.FlowBreak = false;
		this.btnSubmit.Location = new System.Drawing.Point(3, 8);
		this.btnSubmit.Margin = new System.Windows.Forms.Padding(3, 3, 5, 3);
		this.btnSubmit.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -294724347 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6);
		this.btnSubmit.Padding = new System.Windows.Forms.Padding(8, 4, 8, 4);
		this.btnSubmit.Size = new System.Drawing.Size(72, 27);
		this.btnSubmit.TabIndex = 4;
		this.btnSubmit.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA753D1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
		this.btnSubmit.Visible = false;
		this.eLoader1.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.eLoader1.FlowBreak = false;
		this.eLoader1.ForeColor = System.Drawing.Color.DarkBlue;
		this.eLoader1.Location = new System.Drawing.Point(83, 8);
		this.eLoader1.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68D0489C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6);
		this.eLoader1.Size = new System.Drawing.Size(1096, 28);
		this.eLoader1.TabIndex = 5;
		base.Controls.Add(this.top);
		base.Controls.Add(this.eLayer1);
		base.Controls.Add(this.buttomLy);
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1163AA93 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca);
		this.Padding = new System.Windows.Forms.Padding(15, 12, 15, 12);
		base.Size = new System.Drawing.Size(1220, 800);
		this.top.ResumeLayout(false);
		this.top.PerformLayout();
		this.eLayer1.ResumeLayout(false);
		this.body.ResumeLayout(false);
		this.eLayer2.ResumeLayout(false);
		this.buttomLy.ResumeLayout(false);
		this.buttomLy.PerformLayout();
		base.ResumeLayout(false);
	}
}
