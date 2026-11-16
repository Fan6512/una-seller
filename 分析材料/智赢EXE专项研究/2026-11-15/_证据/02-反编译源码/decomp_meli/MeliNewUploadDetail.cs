using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Drawing;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Threading.Tasks;
using System.Windows.Forms;
using Enumer;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Channel;
using ZYing.Data;
using ZYing.Dialog;
using ZYing.Static;
using ZYing.UI;
using ZYing.Web;

namespace ZYing.Detail;

[ToolboxItem(false)]
public class Meli新版上传Detail : ELayer
{
	public Action Submit;

	public Action<ZClient, string> Receiving;

	private IContainer components;

	private EGroup chkGrade;

	private ELine eLine2;

	private EDropLinkageSearch sel站点;

	private ELabelEdit lbl授权;

	private ELine eLine3;

	private ELocalSearch m产品分类;

	private ELine eLine4;

	private EButton btnSubmit;

	private ELabel lblOK;

	private EDropLinkageSearch meli分类;

	private ELabel lbl查找meli;

	private ELine eLine1;

	private EText txtTop;

	private EGroup gType;

	private ELabel lbl历史;

	private ELoader loader;

	private ESelectDateRange selDate;

	private EGroup grpSort;

	private ECombo PLogin;

	private EGroup gLevel;

	private ELabel lblHelp;

	private EPropControls attrs;

	private ESelectText sel保证类型;

	private ECombo sel默认尺寸;

	private EText txtLow;

	private EText txtHigh;

	private EValueText txt采集店铺;

	private EText txtWord;

	private ESelect selArea;

	private ECombo PSource;

	private EText eBrand;

	private ECheck chk编号;

	private ETextArea txtIDs;

	private ECheck 参数补充;

	private EText txtWeight;

	private ETextJoin txtSize;

	private ECheck chk重量过滤;

	private ELine eLine5;

	private EGroup grpAbbrs;

	private EText txt库存;

	private EText txtPrefix;

	private EText txtSuffix;

	private ETextPairs txt替换;

	private ELine eLine6;

	private ECheck chk替换库存;

	private ECheck chk替换尺寸;

	private ELine eLine7;

	private ELayer ly运费;

	private ELabel lb运费;

	private EGroup chk排除平台;

	private EGroup chk允许币种;

	private ELabel lb模板;

	private ECheck chk运费开关;

	private ETabMulti tabLans;

	private ETextAreaMulti txtMlMemo;

	private ETextAreaMulti txtMlTail;

	private ECheck chk不联动分类;

	private ECombo sel尺码表;

	private ECheck chk替换;

	private EGroup selListType;

	private EGroup sel定价模式;

	public static Meli新版上传Detail Instance => Global.Create(typeof(Meli新版上传Detail)) as Meli新版上传Detail;

	public string Ids
	{
		get
		{
			if (!chk编号.Checked)
			{
				return "";
			}
			return txtIDs.Text;
		}
		set
		{
			txtIDs.Text = value;
			chk编号.FireCheck(txtIDs.Text.Length > 0);
		}
	}

	public Meli新版上传Detail()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		txtSize[0].PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B1DBD7B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791);
		txtSize[1].PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E476BFB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		txtSize[2].PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E5B03D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c);
		base.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7642040D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
		lbl授权.Click += delegate
		{
			ZYing.Channel.Meli.Instance.Focus(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76461F69 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835));
		};
		if (!Member.只看自己产品 && (Member.Limit.系统.看业务员 || Member.Limit.系统.改业务员))
		{
			PLogin.Items = Login.Items;
			PLogin.Visible = Login.Items.Count > 0;
		}
		chk编号.CheckedChanged += delegate
		{
			txtIDs.Visible = chk编号.Checked;
		};
		lbl历史.EIcos = IcoEnums.list;
		lbl历史.Click += delegate
		{
			DKindMeli instance = DKindMeli.Instance;
			instance.ChooseAction = (string localid) =>
			{
				m产品分类.Value = localid;
				M产品分类_SelectedChanged(this, EventArgs.Empty);
			};
			instance.Show(sel站点, m产品分类.Id);
		};
		lbl查找meli.EIcos = IcoEnums.vars;
		lbl查找meli.Click += LblFindMeli_Click;
		PSource.Items = Enumer.Web.Items;
		selArea.Items = Enumer.Meli.AccAbbrs;
		sel站点.Items = Enumer.Meli.GroupAccs;
		sel站点.Pop += delegate
		{
			DMeli店铺.Instance.Show(sel站点);
		};
		sel站点.SelectedChanged += 站点_SelectedChanged;
		grpAbbrs.SelectedChanged += GrpAbbrs_SelectedChanged;
		m产品分类.Kinder = KindLocal.产品分类;
		m产品分类.SelectedChanged += M产品分类_SelectedChanged;
		meli分类.SelectedChanged += Meli分类_SelectedChanged;
		meli分类.Pop += LblFindMeli_Click;
		meli分类.Kinder = KindMeli.Instance;
		gType.Items = new EItemCollection(new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E30C752 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B1E9929 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E53A80 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60F8B3D9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5)));
		gType.SelectedChanged += delegate
		{
			chkGrade.Visible = gType.Value == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F2175D4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a);
		};
		gLevel.Items = new EItemCollection(new EItem("", kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -164177204 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd)), Enumer.Sale.Level);
		chkGrade.Items = new EItemCollection(new EItem("", kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9C69CB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947)), Enumer.Sale.Grade);
		grpSort.Items.AddRange(new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x232630E4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5112E801 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D79A2A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6CBF66 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6397C190 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB2B867 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018)));
		grpSort.Value = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643B7D0C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		btnSubmit.Click += BtnSubmit_Click;
		sel默认尺寸.Items = Enumer.Meli.DefaultSizes;
		sel默认尺寸.SelectedChanged += Sel默认尺寸_SelectedChanged;
		sel默认尺寸.SelectedItem = sel默认尺寸.Items[0];
		chk替换库存.CheckedChanged += delegate
		{
			txt库存.ReadOnly = !chk替换库存.Checked;
		};
		selListType.Items = new EItemCollection(new EItem("", kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5014A6A0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408FE96C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63630B3F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0xC090520 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -902805885 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273)));
		selListType.SelectedItem = selListType.Items[0];
		sel定价模式.Items = new EItemCollection(new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x511070CB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB463CD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F211036 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x20063DA8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f), Color.DarkRed));
		sel定价模式.SelectedChanged += delegate
		{
			bool flag = Valid.ToBool(sel定价模式.Value);
			ly运费.Visible = !flag;
			if (flag)
			{
				chk运费开关.Checked = false;
			}
		};
		sel定价模式.Value = (Member.Config.MeliNot_NetProceeds ? kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1164344D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca) : kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3AEFFC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273));
		sel定价模式.OnSelectedChanged();
		chk运费开关.CheckedChanged += delegate
		{
			chk排除平台.Visible = chk运费开关.Checked;
			chk允许币种.Visible = chk运费开关.Checked;
		};
		lb模板.Click += delegate
		{
			DMeli运费计算.Instance.Show(this);
		};
		chk排除平台.Items = new EItemCollection(new EItem(Enumer.Site.美客多), new EItem(Enumer.Site.亚马逊), new EItem(Enumer.Site.速卖通), new EItem(51, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E59176 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce)), new EItem(Enumer.Site.天猫), new EItem(Enumer.Site.淘宝));
		chk排除平台.Values = new string[2]
		{
			$"{8}",
			$"{1}"
		};
		chk允许币种.Items = new EItemCollection(new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DC832B4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x605A279 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B1C9D83 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6CD21C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x3CFEF435 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACC6326 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -164273778 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51140C63 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -776468909 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C1042B4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F44ADBE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5023420B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D314DC6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A74CD4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010)));
		chk允许币种.Value = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68D06CDA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6);
		sel保证类型.SelectedChanged += delegate
		{
			int serial = sel保证类型.SelectedItem.Serial;
			if (serial > 0)
			{
				sel保证类型.Sub.Visible = true;
				sel保证类型.Sub.NeedText = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E1D61C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce);
				sel保证类型.Sub.Text = serial.ToString();
			}
			else
			{
				sel保证类型.Sub.Visible = false;
				sel保证类型.Sub.NeedText = "";
				sel保证类型.Sub.Text = "";
			}
		};
		sel保证类型.Items = Enumer.Meli.Warrantys;
		sel保证类型.SelectedItem = Enumer.Meli.Warrantys.First;
		tabLans.Items = new EItemCollection(new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2003CC1A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x5C5F17BE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E423A2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5827F851 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd)), new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B2968E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x86D3C1A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018)));
		tabLans.SelectedChanged += delegate
		{
			ETextAreaMulti eTextAreaMulti = txtMlMemo;
			string lan = (txtMlTail.Lan = tabLans.Value);
			eTextAreaMulti.Lan = lan;
		};
		sel尺码表.RightClick += async delegate
		{
			if (sel站点.SelectedItem != null)
			{
				await DMeli尺码表.Instance.Show(this, sel站点.SelectedItem.Group, OnSizeChoose);
			}
		};
		if (Member.UserID == 6238)
		{
			txtWeight.Visible = true;
			txtSize.Visible = true;
			chk替换尺寸.Visible = true;
			sel默认尺寸.Visible = true;
		}
	}

	public override void OnActive()
	{
		base.OnActive();
		TakeFocus();
	}

	protected override void OnVisibleChanged(EventArgs e)
	{
		base.OnVisibleChanged(e);
		if (Visible)
		{
			TakeFocus();
		}
	}

	private void TakeFocus()
	{
		if (Visible && !ContainsFocus)
		{
			Focus();
		}
	}

	public override void Load()
	{
		base.Load();
		AppConfig.MeliSubmitPara meliSubmitPara = JsonConfig<AppConfig>.Current.MeliPara ?? new AppConfig.MeliSubmitPara();
		chk重量过滤.Checked = meliSubmitPara.Select_Weight;
		chk运费开关.Checked = meliSubmitPara.Add_Freight_Enable;
		chk运费开关.FireCheck(meliSubmitPara.Add_Freight_Enable);
		if (meliSubmitPara.Add_Freight_Enable)
		{
			chk允许币种.Values = meliSubmitPara.Include_Lans;
			chk排除平台.Values = meliSubmitPara.Except_Sites;
		}
		chk不联动分类.Checked = meliSubmitPara.NotLoadCates;
	}

	private void Sel默认尺寸_SelectedChanged(object sender, EventArgs e)
	{
		EItem selectedItem = sel默认尺寸.SelectedItem;
		if (selectedItem == null || selectedItem.Extra.IsNullOrEmpty())
		{
			txtWeight.Text = "";
			txtSize.Text = "";
			return;
		}
		int[] array = Split.Ints(selectedItem.Extra);
		if (array.Length == 3)
		{
			txtSize.Ints = new int[3]
			{
				array[0],
				array[1],
				array[2]
			};
			int[] array2 = Split.Ints(selectedItem.Key, '-');
			txtWeight.Int = Rnd.Next(array2[0], array2[1]);
		}
	}

	private async void OnSizeChoose(string sizeid)
	{
		int num = sel站点.SelectedItem?.Group ?? 0;
		if (num != 0)
		{
			ECombo eCombo = sel尺码表;
			eCombo.Items = await MeliSizeChart.GetItems(num);
			sel尺码表.Value = sizeid;
		}
	}

	private async void 站点_SelectedChanged(object sender, EventArgs e)
	{
		EItem it = sel站点.SelectedItem;
		if (it == null)
		{
			grpAbbrs.Visible = false;
			return;
		}
		EDropLinkageSearch eDropLinkageSearch = meli分类;
		ELocalSearch eLocalSearch = m产品分类;
		bool flag = (sel尺码表.Enabled = true);
		bool enabled = (eLocalSearch.Enabled = flag);
		eDropLinkageSearch.Enabled = enabled;
		ELabel eLabel = lbl历史;
		enabled = (lbl查找meli.Enabled = true);
		eLabel.Enabled = enabled;
		if (it.SubItems.Count > 0)
		{
			grpAbbrs.Items.Clear();
			grpAbbrs.Items.Add(new EItem("", kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E45F602 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee)));
			foreach (EItem subItem in it.SubItems)
			{
				EItem eItem = subItem.Clone();
				eItem.Color = Color.Empty;
				grpAbbrs.Items.Add(eItem);
			}
			grpAbbrs.Visible = true;
			grpAbbrs.Values = ReadAbbrs(grpAbbrs.Items);
		}
		else
		{
			grpAbbrs.Items.Clear();
			grpAbbrs.Items.Add(it.Clone());
			grpAbbrs.Values = new string[1] { it.Key };
			grpAbbrs.Visible = false;
		}
		await LoadKind();
		sel尺码表.Reset();
		ECombo eCombo = sel尺码表;
		eCombo.Items = await MeliSizeChart.GetItems(it.Group);
	}

	private void GrpAbbrs_SelectedChanged(object sender, EventArgs e)
	{
		SavePara();
	}

	private string[] ReadAbbrs(EItemCollection items)
	{
		AppConfig.MeliSubmitPara meliSubmitPara = JsonConfig<AppConfig>.Current.MeliPara ?? new AppConfig.MeliSubmitPara();
		if (!meliSubmitPara.AbbrNames.IsEmpty())
		{
			List<string> list = new List<string>();
			foreach (EItem item in items)
			{
				if (((IEnumerable<string>)meliSubmitPara.AbbrNames).Contains(new string[1] { item.Text }))
				{
					list.Add(item.Key);
				}
			}
			return list.ToArray();
		}
		return (from x in items
			where !Valid.ToBool(x.Tag)
			select x.Key).ToArray();
	}

	private async void M产品分类_SelectedChanged(object sender, EventArgs e)
	{
		if (sel站点.SelectedItem != null)
		{
			if (m产品分类.Value.IsNullOrEmpty())
			{
				meli分类.Set("");
				attrs.Text = "";
			}
			else
			{
				await LoadKind();
			}
		}
	}

	private async Task LoadKind()
	{
		if (chk不联动分类.Checked)
		{
			return;
		}
		EItem selectedItem = sel站点.SelectedItem;
		if (selectedItem == null)
		{
			return;
		}
		meli分类.Tag = null;
		int value = selectedItem.Group;
		int value2 = 0;
		if (selectedItem.SubCount == 0)
		{
			value2 = selectedItem.Id;
		}
		Record record = new Record();
		record.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B1CB2E7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791)).Set(m产品分类.Value);
		record.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16358DE1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8)).Set(value2);
		record.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0E0C39 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd)).Set(value);
		ZClient zc = ZClient.Create(ZClient.美客多);
		zc.Loader = loader;
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A876170 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21);
		await zc.PostAsync(record);
		Record record2 = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9D7D17 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947)];
		if (record2.Count != 0)
		{
			int num = record2.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D37AAB4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b));
			meli分类.Set((num > 0) ? num.ToString() : m产品分类.Value);
			string text = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468FF34A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6));
			if (!text.IsNullOrEmpty() && text != kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B7DEF6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2))
			{
				meli分类.Tag = text;
			}
			int num2 = record2.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37BA1D45 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e));
			if (num2 > 0)
			{
				sel保证类型.Value = num2.ToString();
			}
			int num3 = record2.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA2A51A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b));
			if (num3 > 0)
			{
				sel保证类型.Text = num3.ToString();
			}
			string text2 = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C145BEA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1));
			if (!text2.IsNullOrEmpty())
			{
				eBrand.Text = text2;
			}
			txtPrefix.Text = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -59787110 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1));
			txtSuffix.Text = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE45168 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde));
			txtMlMemo.Reset();
			txtMlTail.Reset();
			txtMlMemo.Multi = new NMulti(record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x200221F0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f)));
			txtMlTail.Multi = new NMulti(record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527FA1BB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1)));
		}
	}

	private void LblFindMeli_Click(object sender, EventArgs e)
	{
		if (sel站点.SelectedItem == null)
		{
			throw new ApplicationException(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60F8B3A7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5));
		}
		DMeli分类.Instance.Show(meli分类);
	}

	private void Meli分类_SelectedChanged(object sender, EventArgs e)
	{
		EItem selectedItem = meli分类.SelectedItem;
		if (selectedItem == null)
		{
			meli分类.Hint = "";
			attrs.Props = null;
			eBrand.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B1707C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
			eBrand.Visible = true;
			return;
		}
		string abbr = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA827EC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b);
		List<NAttr> list = attrs.Attrs;
		attrs.Props = EPropHelper.GetMeliProp(abbr).GetProps(int.Parse(selectedItem.Key));
		if (meli分类.Tag != null && meli分类.Tag is string json)
		{
			attrs.Attrs = Json.ReadList<NAttr>(json);
		}
		else
		{
			attrs.Attrs = list;
		}
		if ((attrs.Props?.Props?.FirstOrDefault((EProp x) => x.Key == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AAE8AA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611)))?.Values != null)
		{
			eBrand.Text = "";
		}
		meli分类.Hint = selectedItem.GetFullSub(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x867CB41 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788));
	}

	private async void BtnSubmit_Click(object sender, EventArgs e)
	{
		Record record = GetRecord();
		int[] array = grpAbbrs.Values.Select((string x) => x.ToInt()).ToArray();
		if (array.Length == 0)
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E397356 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6));
		}
		record.Columns.AddInts(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A907568 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380)).Set(array);
		if (chk编号.Checked)
		{
			int[] array2 = Split.Ints(txtIDs.Text);
			if (array2.Length == 0)
			{
				throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA51969B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1));
			}
			DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(15, 1);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x501950FC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e));
			defaultInterpolatedStringHandler.AppendFormatted(array2.Length);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60F7BBB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2));
			if (!EDialog.Confirm(defaultInterpolatedStringHandler.ToStringAndClear()))
			{
				return;
			}
		}
		else
		{
			if (chkGrade.Visible && chkGrade.SelectedItems.Count == 0)
			{
				throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -7034546 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552));
			}
			record.Columns.AddInts(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16355853 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8)).Set(chkGrade.Values);
			if (m产品分类.Value.IsNullOrEmpty() && !EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E4F62CF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689)))
			{
				return;
			}
		}
		if (chk替换尺寸.Checked)
		{
			int num = txtWeight.Int;
			int[] ints = txtSize.Ints;
			if (num < 10 || ints.IsEmpty() || ints.Any((int x) => x < 2))
			{
				throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x66C937E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8));
			}
		}
		if (!selDate.Start.Value.IsNullOrEmpty())
		{
			record.Columns.AddDate(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BCB443 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef)).Set(selDate.Start.Value);
		}
		if (!selDate.End.Value.IsNullOrEmpty())
		{
			record.Columns.AddDate(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB7505D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4)).Set(selDate.End.Value);
		}
		record.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63F008A1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1)).Set(attrs.Json);
		record.Columns.AddInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643B0E36 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848)).Set(chk替换库存.Checked ? txt库存.Int : 0);
		record.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B161551 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791)).Set(sel尺码表.Value);
		bool flag = Valid.ToBool(sel定价模式.Value);
		if (flag || EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA523713 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1)))
		{
			record.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -1894522442 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835)).Set(flag);
			if (chk运费开关.Checked)
			{
				record.Columns.AddBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22406EB0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2)).Set(value: true);
				record.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x582C60CD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd)).Set(chk排除平台.Values);
				record.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x66FA7F2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8)).Set(chk允许币种.Values);
			}
			if (!txtPrefix.Text.IsNullOrEmpty() && txtPrefix.Text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x322E4E4E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594)))
			{
				throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x636D5DD5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3));
			}
			if (!txtSuffix.Text.IsNullOrEmpty() && txtSuffix.Text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22406ED0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2)))
			{
				throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -2135048172 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97));
			}
			Submit?.Invoke();
			ZClient zClient = ZClient.Create(ZClient.美客多2);
			zClient.UI = Parent;
			zClient.Button = btnSubmit;
			zClient.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -950718172 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8);
			if (Receiving != null)
			{
				zClient.Receiving = Receiving;
			}
			zClient.Failed = (Exception ex) =>
			{
				EDialog.Show(ex);
			};
			await zClient.PostAsync(record);
			lblOK.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EAE5BAB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
			SavePara();
		}
	}

	protected override bool ProcessCmdKey(ref Message msg, Keys keyData)
	{
		if (keyData == Keys.F6)
		{
			if (btnSubmit.Enabled)
			{
				BtnSubmit_Click(btnSubmit, EventArgs.Empty);
			}
			return true;
		}
		return base.ProcessCmdKey(ref msg, keyData);
	}

	private void SavePara()
	{
		AppConfig.MeliSubmitPara meliSubmitPara = new AppConfig.MeliSubmitPara();
		meliSubmitPara.Select_Weight = chk重量过滤.Checked;
		meliSubmitPara.Add_Freight_Enable = chk运费开关.Checked;
		if (chk运费开关.Checked)
		{
			meliSubmitPara.Include_Lans = chk允许币种.Values;
			meliSubmitPara.Except_Sites = chk排除平台.Values;
		}
		meliSubmitPara.AbbrNames = grpAbbrs.SelectedItems?.Select((EItem x) => x.Text).ToArray();
		meliSubmitPara.NotLoadCates = chk不联动分类.Checked;
		JsonConfig<AppConfig>.Current.MeliPara = meliSubmitPara;
		JsonConfig<AppConfig>.Current.Save();
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
		ZYing.NMulti multi = new ZYing.NMulti();
		ZYing.NMulti multi2 = new ZYing.NMulti();
		this.lblHelp = new ZYing.UI.ELabel();
		this.gType = new ZYing.UI.EGroup();
		this.gLevel = new ZYing.UI.EGroup();
		this.chkGrade = new ZYing.UI.EGroup();
		this.chk编号 = new ZYing.UI.ECheck();
		this.txtIDs = new ZYing.UI.ETextArea();
		this.txtTop = new ZYing.UI.EText();
		this.selDate = new ZYing.UI.ESelectDateRange();
		this.PLogin = new ZYing.UI.ECombo();
		this.txtLow = new ZYing.UI.EText();
		this.txtHigh = new ZYing.UI.EText();
		this.PSource = new ZYing.UI.ECombo();
		this.selArea = new ZYing.UI.ESelect();
		this.txt采集店铺 = new ZYing.UI.EValueText();
		this.txtWord = new ZYing.UI.EText();
		this.grpSort = new ZYing.UI.EGroup();
		this.eLine1 = new ZYing.UI.ELine();
		this.sel站点 = new ZYing.UI.EDropLinkageSearch();
		this.lbl授权 = new ZYing.UI.ELabelEdit();
		this.loader = new ZYing.UI.ELoader();
		this.grpAbbrs = new ZYing.UI.EGroup();
		this.eLine2 = new ZYing.UI.ELine();
		this.lbl历史 = new ZYing.UI.ELabel();
		this.meli分类 = new ZYing.UI.EDropLinkageSearch();
		this.lbl查找meli = new ZYing.UI.ELabel();
		this.eLine3 = new ZYing.UI.ELine();
		this.参数补充 = new ZYing.UI.ECheck();
		this.chk重量过滤 = new ZYing.UI.ECheck();
		this.eBrand = new ZYing.UI.EText();
		this.sel保证类型 = new ZYing.UI.ESelectText();
		this.txtWeight = new ZYing.UI.EText();
		this.txtSize = new ZYing.UI.ETextJoin();
		this.sel默认尺寸 = new ZYing.UI.ECombo();
		this.eLine4 = new ZYing.UI.ELine();
		this.btnSubmit = new ZYing.UI.EButton();
		this.lblOK = new ZYing.UI.ELabel();
		this.eLine5 = new ZYing.UI.ELine();
		this.txt库存 = new ZYing.UI.EText();
		this.txtPrefix = new ZYing.UI.EText();
		this.txtSuffix = new ZYing.UI.EText();
		this.txt替换 = new ZYing.UI.ETextPairs();
		this.eLine6 = new ZYing.UI.ELine();
		this.chk替换库存 = new ZYing.UI.ECheck();
		this.chk替换尺寸 = new ZYing.UI.ECheck();
		this.eLine7 = new ZYing.UI.ELine();
		this.ly运费 = new ZYing.UI.ELayer();
		this.lb运费 = new ZYing.UI.ELabel();
		this.chk运费开关 = new ZYing.UI.ECheck();
		this.lb模板 = new ZYing.UI.ELabel();
		this.chk允许币种 = new ZYing.UI.EGroup();
		this.chk排除平台 = new ZYing.UI.EGroup();
		this.tabLans = new ZYing.UI.ETabMulti();
		this.txtMlMemo = new ZYing.UI.ETextAreaMulti();
		this.txtMlTail = new ZYing.UI.ETextAreaMulti();
		this.chk不联动分类 = new ZYing.UI.ECheck();
		this.attrs = new ZYing.UI.EPropControls();
		this.sel尺码表 = new ZYing.UI.ECombo();
		this.chk替换 = new ZYing.UI.ECheck();
		this.selListType = new ZYing.UI.EGroup();
		this.sel定价模式 = new ZYing.UI.EGroup();
		this.m产品分类 = new ZYing.UI.ELocalSearch();
		this.ly运费.SuspendLayout();
		base.SuspendLayout();
		this.lblHelp.AllowDrop = true;
		this.lblHelp.AutoSize = false;
		this.lblHelp.Blink = true;
		this.lblHelp.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.lblHelp.Fonts = ZYing.UI.EFonts.微软雅黑Bold;
		this.lblHelp.ForeColor = System.Drawing.Color.DarkRed;
		this.lblHelp.Location = new System.Drawing.Point(3, 11);
		this.lblHelp.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E307788 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.lblHelp.Padding = new System.Windows.Forms.Padding(0, 4, 0, 4);
		this.lblHelp.Size = new System.Drawing.Size(928, 27);
		this.lblHelp.TabIndex = 389;
		this.lblHelp.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D297088 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3);
		this.gType.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B00A51 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e);
		this.gType.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.gType.Indent = 76;
		this.gType.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AF2918 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611);
		this.gType.Location = new System.Drawing.Point(3, 44);
		this.gType.MaxSelect = 1;
		this.gType.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE4A4E0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde);
		this.gType.NeedText = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-2066591320) ^ 0x11A1E1B2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21);
		this.gType.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.gType.Size = new System.Drawing.Size(928, 28);
		this.gType.TabIndex = 237;
		this.gLevel.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B17104 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.gLevel.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.gLevel.Indent = 76;
		this.gLevel.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E73712 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce);
		this.gLevel.Location = new System.Drawing.Point(3, 78);
		this.gLevel.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA55387 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
		this.gLevel.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.gLevel.Size = new System.Drawing.Size(928, 28);
		this.gLevel.TabIndex = 318;
		this.chkGrade.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9C1C0B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
		this.chkGrade.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.chkGrade.Indent = 76;
		this.chkGrade.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9E40E5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
		this.chkGrade.Location = new System.Drawing.Point(3, 112);
		this.chkGrade.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63959842 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2);
		this.chkGrade.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.chkGrade.Size = new System.Drawing.Size(928, 28);
		this.chkGrade.TabIndex = 1;
		this.chk编号.Indent = 76;
		this.chk编号.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x105195CE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf);
		this.chk编号.Location = new System.Drawing.Point(3, 146);
		this.chk编号.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A387B98 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273);
		this.chk编号.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk编号.Size = new System.Drawing.Size(419, 28);
		this.chk编号.TabIndex = 518;
		this.chk编号.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5012715C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e);
		this.txtIDs.AcceptReturn = false;
		this.txtIDs.Always = true;
		this.txtIDs.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.txtIDs.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA582C9D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1);
		this.txtIDs.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.txtIDs.Indent = 76;
		this.txtIDs.Location = new System.Drawing.Point(3, 178);
		this.txtIDs.Margin = new System.Windows.Forms.Padding(3, 1, 3, 3);
		this.txtIDs.MaximumSize = new System.Drawing.Size(10000, 207);
		this.txtIDs.MaxLength = 4000;
		this.txtIDs.MaxLines = 10;
		this.txtIDs.MinimumSize = new System.Drawing.Size(0, 47);
		this.txtIDs.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACEF798 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057);
		this.txtIDs.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtIDs.PostVisible = true;
		this.txtIDs.Size = new System.Drawing.Size(928, 47);
		this.txtIDs.TabIndex = 517;
		this.txtIDs.TabStop = false;
		this.txtIDs.Visible = false;
		this.txtIDs.VisibleDesign = false;
		this.txtTop.Always = true;
		this.txtTop.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.txtTop.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA594317 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1);
		this.txtTop.FixWidth = 185;
		this.txtTop.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.txtTop.Indent = 76;
		this.txtTop.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5112E97F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
		this.txtTop.Location = new System.Drawing.Point(3, 231);
		this.txtTop.MaxLength = 4;
		this.txtTop.MaxValue = 200;
		this.txtTop.MinValue = 1;
		this.txtTop.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E73760 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce);
		this.txtTop.NeedText = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x19B1CA4F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.txtTop.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtTop.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x639598F8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2);
		this.txtTop.Size = new System.Drawing.Size(928, 28);
		this.txtTop.TabIndex = 228;
		this.txtTop.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F21A414 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a);
		this.txtTop.ValidType = ZYing.UI.EValidType.Number;
		this.selDate.Always = true;
		this.selDate.EndField = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51101267 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
		this.selDate.EndPlaceholder = "";
		this.selDate.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x58249765 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd);
		this.selDate.Indent = 76;
		this.selDate.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C16A714 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.selDate.Location = new System.Drawing.Point(3, 265);
		this.selDate.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB686D1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4);
		this.selDate.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.selDate.Size = new System.Drawing.Size(300, 28);
		this.selDate.StartField = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E46E37 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c);
		this.selDate.StartPlaceholder = "";
		this.selDate.TabIndex = 293;
		this.PLogin.Always = true;
		this.PLogin.AutoHeight = true;
		this.PLogin.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63970C6E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2);
		this.PLogin.FixWidth = 300;
		this.PLogin.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.PLogin.Indent = 76;
		this.PLogin.IsExact = true;
		this.PLogin.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5025A7B7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841);
		this.PLogin.Location = new System.Drawing.Point(3, 299);
		this.PLogin.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A8EB5B6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21);
		this.PLogin.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PLogin.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x606B689 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2);
		this.PLogin.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C98A51B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
		this.PLogin.Size = new System.Drawing.Size(928, 28);
		this.PLogin.TabIndex = 310;
		this.txtLow.Always = true;
		this.txtLow.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.txtLow.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B1543B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		this.txtLow.FlowBreak = false;
		this.txtLow.Indent = 76;
		this.txtLow.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5025A46F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841);
		this.txtLow.Location = new System.Drawing.Point(3, 333);
		this.txtLow.Margin = new System.Windows.Forms.Padding(3, 3, 1, 3);
		this.txtLow.MaxLength = 4;
		this.txtLow.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-713251553) ^ 0x49E6886A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3);
		this.txtLow.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtLow.Size = new System.Drawing.Size(180, 28);
		this.txtLow.TabIndex = 476;
		this.txtLow.ValidType = ZYing.UI.EValidType.Price;
		this.txtHigh.Always = true;
		this.txtHigh.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.txtHigh.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B17230 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.txtHigh.Indent = 12;
		this.txtHigh.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B1E9B65 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791);
		this.txtHigh.Location = new System.Drawing.Point(184, 333);
		this.txtHigh.Margin = new System.Windows.Forms.Padding(0, 3, 3, 3);
		this.txtHigh.MaxLength = 4;
		this.txtHigh.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x66728E8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8);
		this.txtHigh.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtHigh.Size = new System.Drawing.Size(119, 28);
		this.txtHigh.TabIndex = 477;
		this.txtHigh.ValidType = ZYing.UI.EValidType.Price;
		this.PSource.AutoHeight = true;
		this.PSource.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B346FF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		this.PSource.FlowBreak = false;
		this.PSource.Indent = 76;
		this.PSource.IsExact = true;
		this.PSource.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63959554 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2);
		this.PSource.Location = new System.Drawing.Point(3, 367);
		this.PSource.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B17C30 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.PSource.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PSource.Size = new System.Drawing.Size(300, 28);
		this.PSource.TabIndex = 495;
		this.selArea.AutoHeight = true;
		this.selArea.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x6CCDAAD0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.selArea.FixWidth = 100;
		this.selArea.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.selArea.Location = new System.Drawing.Point(309, 367);
		this.selArea.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527F50C1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1);
		this.selArea.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.selArea.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D6D3BD0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9);
		this.selArea.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5824E1EF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd);
		this.selArea.Size = new System.Drawing.Size(622, 28);
		this.selArea.TabIndex = 489;
		this.txt采集店铺.AutoHeight = true;
		this.txt采集店铺.FixWidth = 406;
		this.txt采集店铺.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.txt采集店铺.Indent = 76;
		this.txt采集店铺.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A57B7CA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39);
		this.txt采集店铺.Location = new System.Drawing.Point(3, 401);
		this.txt采集店铺.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -1874752444 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		this.txt采集店铺.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txt采集店铺.ReadOnly = true;
		this.txt采集店铺.RightEIcos = ZYing.IcoEnums.edit_hover;
		this.txt采集店铺.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x8664627 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		this.txt采集店铺.Size = new System.Drawing.Size(928, 28);
		this.txt采集店铺.TabIndex = 478;
		this.txt采集店铺.TabStop = false;
		this.txt采集店铺.Visible = false;
		this.txtWord.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.txtWord.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x20033FDC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f);
		this.txtWord.FixWidth = 406;
		this.txtWord.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.txtWord.Indent = 76;
		this.txtWord.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B292C9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e);
		this.txtWord.Location = new System.Drawing.Point(3, 435);
		this.txtWord.MaxLength = 50;
		this.txtWord.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE562E2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde);
		this.txtWord.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtWord.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60F8B1AF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5);
		this.txtWord.Size = new System.Drawing.Size(928, 28);
		this.txtWord.TabIndex = 483;
		this.grpSort.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA594CF7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1);
		this.grpSort.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.grpSort.Indent = 76;
		this.grpSort.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D37AB36 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
		this.grpSort.Location = new System.Drawing.Point(3, 503);
		this.grpSort.MaxSelect = 1;
		this.grpSort.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x53B57726 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5);
		this.grpSort.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.grpSort.Size = new System.Drawing.Size(928, 28);
		this.grpSort.TabIndex = 299;
		this.eLine1.Location = new System.Drawing.Point(3, 539);
		this.eLine1.Margin = new System.Windows.Forms.Padding(3, 5, 3, 5);
		this.eLine1.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DC8BF98 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5);
		this.eLine1.Size = new System.Drawing.Size(928, 2);
		this.sel站点.AutoHeight = true;
		this.sel站点.FixWidth = 406;
		this.sel站点.FlowBreak = false;
		this.sel站点.Indent = 76;
		this.sel站点.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E317A0A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.sel站点.Location = new System.Drawing.Point(3, 549);
		this.sel站点.MaxHeight = -24;
		this.sel站点.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C1CD84 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.sel站点.NeedText = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AC0C0C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611);
		this.sel站点.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.sel站点.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B1B57D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e);
		this.sel站点.Size = new System.Drawing.Size(408, 28);
		this.sel站点.TabIndex = 4;
		this.lbl授权.FlowBreak = false;
		this.lbl授权.Location = new System.Drawing.Point(417, 550);
		this.lbl授权.Margin = new System.Windows.Forms.Padding(3, 4, 3, 3);
		this.lbl授权.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D46FBE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda);
		this.lbl授权.Size = new System.Drawing.Size(63, 28);
		this.lbl授权.TabIndex = 170;
		this.lbl授权.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C1479C8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.loader.FlowBreak = false;
		this.loader.Location = new System.Drawing.Point(486, 549);
		this.loader.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB75F41 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4);
		this.loader.Size = new System.Drawing.Size(308, 28);
		this.loader.TabIndex = 261;
		this.loader.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60F8C91D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5);
		this.grpAbbrs.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.grpAbbrs.Indent = 76;
		this.grpAbbrs.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC7CC28 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332);
		this.grpAbbrs.Location = new System.Drawing.Point(3, 584);
		this.grpAbbrs.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3226B836 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		this.grpAbbrs.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.grpAbbrs.Size = new System.Drawing.Size(928, 28);
		this.grpAbbrs.TabIndex = 560;
		this.grpAbbrs.Visible = false;
		this.eLine2.Location = new System.Drawing.Point(3, 620);
		this.eLine2.Margin = new System.Windows.Forms.Padding(3, 5, 3, 5);
		this.eLine2.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA8B002 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b);
		this.eLine2.Size = new System.Drawing.Size(928, 2);
		this.lbl历史.Enabled = false;
		this.lbl历史.Location = new System.Drawing.Point(895, 631);
		this.lbl历史.Margin = new System.Windows.Forms.Padding(3, 4, 3, 3);
		this.lbl历史.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -2139597878 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
		this.lbl历史.Padding = new System.Windows.Forms.Padding(6, 4, 1, 4);
		this.lbl历史.Size = new System.Drawing.Size(36, 28);
		this.lbl历史.TabIndex = 257;
		this.lbl历史.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C2E884 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.meli分类.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.meli分类.AutoHeight = true;
		this.meli分类.Enabled = false;
		this.meli分类.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5025A571 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841);
		this.meli分类.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.meli分类.FlowBreak = false;
		this.meli分类.FlowWrap = false;
		this.meli分类.Indent = 76;
		this.meli分类.IsFinal = true;
		this.meli分类.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6367DC37 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3);
		this.meli分类.Location = new System.Drawing.Point(3, 665);
		this.meli分类.MaximumSize = new System.Drawing.Size(900, 0);
		this.meli分类.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E4601A6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee);
		this.meli分类.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.meli分类.Size = new System.Drawing.Size(886, 28);
		this.meli分类.TabIndex = 207;
		this.lbl查找meli.Enabled = false;
		this.lbl查找meli.Location = new System.Drawing.Point(895, 666);
		this.lbl查找meli.Margin = new System.Windows.Forms.Padding(3, 4, 3, 3);
		this.lbl查找meli.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7644DC61 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
		this.lbl查找meli.Padding = new System.Windows.Forms.Padding(6, 4, 1, 4);
		this.lbl查找meli.Size = new System.Drawing.Size(36, 28);
		this.lbl查找meli.TabIndex = 209;
		this.lbl查找meli.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F41BDD6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835);
		this.eLine3.Location = new System.Drawing.Point(3, 728);
		this.eLine3.Margin = new System.Windows.Forms.Padding(3, 5, 3, 5);
		this.eLine3.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A923172 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380);
		this.eLine3.Size = new System.Drawing.Size(928, 2);
		this.参数补充.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AF529E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611);
		this.参数补充.Indent = 76;
		this.参数补充.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51129379 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
		this.参数补充.Location = new System.Drawing.Point(3, 738);
		this.参数补充.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16377EEF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8);
		this.参数补充.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.参数补充.Size = new System.Drawing.Size(398, 28);
		this.参数补充.TabIndex = 525;
		this.参数补充.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51129343 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
		this.参数补充.Visible = false;
		this.chk重量过滤.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7644A4FF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
		this.chk重量过滤.Indent = 76;
		this.chk重量过滤.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CF5197 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		this.chk重量过滤.Location = new System.Drawing.Point(3, 469);
		this.chk重量过滤.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1037AD00 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
		this.chk重量过滤.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk重量过滤.Size = new System.Drawing.Size(322, 28);
		this.chk重量过滤.TabIndex = 531;
		this.chk重量过滤.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6A5BBE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.eBrand.Always = true;
		this.eBrand.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.eBrand.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x606ABC3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2);
		this.eBrand.Indent = 76;
		this.eBrand.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C16AA9C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.eBrand.Location = new System.Drawing.Point(3, 806);
		this.eBrand.MaxLength = 50;
		this.eBrand.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x6C68D224 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4);
		this.eBrand.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.eBrand.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0D176F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.eBrand.Size = new System.Drawing.Size(560, 28);
		this.eBrand.TabIndex = 501;
		this.eBrand.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CF2A5D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		this.sel保证类型.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-813636989) ^ 0x15D85021 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010);
		this.sel保证类型.FlowBreak = false;
		this.sel保证类型.Indent = 76;
		this.sel保证类型.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75C9CF01 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		this.sel保证类型.Location = new System.Drawing.Point(3, 1089);
		this.sel保证类型.MaxLength = 3;
		this.sel保证类型.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75C9CF13 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		this.sel保证类型.NeedText = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E34BAF2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.sel保证类型.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.sel保证类型.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x8628C9F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		this.sel保证类型.RootWidth = 260;
		this.sel保证类型.Size = new System.Drawing.Size(461, 28);
		this.sel保证类型.TabIndex = 399;
		this.sel保证类型.ValidType = ZYing.UI.EValidType.Number;
		this.txtWeight.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.txtWeight.Bold = true;
		this.txtWeight.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x200244BE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f);
		this.txtWeight.FlowBreak = false;
		this.txtWeight.Fonts = ZYing.UI.EFonts.VerdanaBold;
		this.txtWeight.Indent = 76;
		this.txtWeight.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -494597459 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		this.txtWeight.Location = new System.Drawing.Point(3, 874);
		this.txtWeight.MaxHeight = -165;
		this.txtWeight.MaxLength = 5;
		this.txtWeight.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D35AC98 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
		this.txtWeight.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtWeight.Size = new System.Drawing.Size(150, 28);
		this.txtWeight.TabIndex = 528;
		this.txtWeight.Unit = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB7AF2D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4);
		this.txtWeight.ValidType = ZYing.UI.EValidType.Number;
		this.txtWeight.Visible = false;
		this.txtSize.ChildNumber = 3;
		this.txtSize.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x70869422 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834);
		this.txtSize.FixWidth = 200;
		this.txtSize.FlowBreak = false;
		this.txtSize.Indent = 40;
		this.txtSize.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E300108 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.txtSize.Location = new System.Drawing.Point(159, 874);
		this.txtSize.Margin = new System.Windows.Forms.Padding(3, 3, 0, 3);
		this.txtSize.MaxLength = 3;
		this.txtSize.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B37548 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.txtSize.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtSize.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x200357D4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f);
		this.txtSize.Size = new System.Drawing.Size(242, 28);
		this.txtSize.SpaceLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACE0B2A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057);
		this.txtSize.TabIndex = 529;
		this.txtSize.TextAlign = System.Windows.Forms.HorizontalAlignment.Center;
		this.txtSize.ValidType = ZYing.UI.EValidType.Number;
		this.txtSize.Visible = false;
		this.sel默认尺寸.AutoHeight = true;
		this.sel默认尺寸.Indent = 76;
		this.sel默认尺寸.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16342179 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8);
		this.sel默认尺寸.Location = new System.Drawing.Point(3, 908);
		this.sel默认尺寸.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FB96C5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5);
		this.sel默认尺寸.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.sel默认尺寸.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A931F26 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380);
		this.sel默认尺寸.Size = new System.Drawing.Size(386, 28);
		this.sel默认尺寸.TabIndex = 402;
		this.sel默认尺寸.Visible = false;
		this.eLine4.Location = new System.Drawing.Point(3, 1123);
		this.eLine4.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B31DF0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.eLine4.Size = new System.Drawing.Size(928, 2);
		this.btnSubmit.FlowBreak = false;
		this.btnSubmit.ForeColor = System.Drawing.Color.Black;
		this.btnSubmit.Location = new System.Drawing.Point(79, 1359);
		this.btnSubmit.Margin = new System.Windows.Forms.Padding(79, 8, 3, 3);
		this.btnSubmit.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5826EC9B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b0d18d37aaf84fdaaa26521091ff83fd);
		this.btnSubmit.Size = new System.Drawing.Size(96, 25);
		this.btnSubmit.TabIndex = 17;
		this.btnSubmit.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68A9BA3C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611);
		this.lblOK.Location = new System.Drawing.Point(162, 1359);
		this.lblOK.Margin = new System.Windows.Forms.Padding(8, 8, 3, 3);
		this.lblOK.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63979F18 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2);
		this.lblOK.Padding = new System.Windows.Forms.Padding(0, 4, 0, 4);
		this.lblOK.Size = new System.Drawing.Size(371, 28);
		this.lblOK.TabIndex = 198;
		this.lblOK.TabStop = false;
		this.eLine5.Location = new System.Drawing.Point(3, 3);
		this.eLine5.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x664DBC6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8);
		this.eLine5.Size = new System.Drawing.Size(928, 2);
		this.txt库存.Always = true;
		this.txt库存.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C16FF8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.txt库存.FlowBreak = false;
		this.txt库存.Indent = 76;
		this.txt库存.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A91FA82 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380);
		this.txt库存.Location = new System.Drawing.Point(3, 840);
		this.txt库存.MinValue = 1;
		this.txt库存.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACDCEF4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057);
		this.txt库存.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txt库存.ReadOnly = true;
		this.txt库存.Size = new System.Drawing.Size(150, 28);
		this.txt库存.TabIndex = 566;
		this.txt库存.TabStop = false;
		this.txt库存.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C143E32 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		this.txt库存.ValidType = ZYing.UI.EValidType.Number;
		this.txtPrefix.Always = true;
		this.txtPrefix.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.txtPrefix.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6E41E0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.txtPrefix.FlowBreak = false;
		this.txtPrefix.FlowWrap = false;
		this.txtPrefix.Indent = 76;
		this.txtPrefix.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1705051699 ^ -623561451 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44);
		this.txtPrefix.Location = new System.Drawing.Point(3, 1165);
		this.txtPrefix.MaxLength = 50;
		this.txtPrefix.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F40B70E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835);
		this.txtPrefix.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtPrefix.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E448470 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee);
		this.txtPrefix.RightColor = System.Drawing.Color.DarkRed;
		this.txtPrefix.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA5C5EEB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1);
		this.txtPrefix.Size = new System.Drawing.Size(549, 28);
		this.txtPrefix.TabIndex = 568;
		this.txtSuffix.Anchor = System.Windows.Forms.AnchorStyles.Top | System.Windows.Forms.AnchorStyles.Left | System.Windows.Forms.AnchorStyles.Right;
		this.txtSuffix.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6397678A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2);
		this.txtSuffix.FixWidth = 225;
		this.txtSuffix.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.txtSuffix.Location = new System.Drawing.Point(558, 1165);
		this.txtSuffix.MaxLength = 50;
		this.txtSuffix.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B1C6621 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791);
		this.txtSuffix.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtSuffix.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2D6B18 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3);
		this.txtSuffix.RightColor = System.Drawing.Color.DarkRed;
		this.txtSuffix.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A51508E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39);
		this.txtSuffix.Size = new System.Drawing.Size(373, 28);
		this.txtSuffix.TabIndex = 569;
		this.txt替换.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224A1C9A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2);
		this.txt替换.Indent = 76;
		this.txt替换.LeftHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B3A977 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		this.txt替换.Location = new System.Drawing.Point(3, 1199);
		this.txt替换.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2D6B4E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3);
		this.txt替换.NeedCheck = true;
		this.txt替换.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txt替换.RightHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B06FCB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e);
		this.txt替换.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A1AB30 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010);
		this.txt替换.Size = new System.Drawing.Size(612, 28);
		this.txt替换.TabIndex = 571;
		this.eLine6.Location = new System.Drawing.Point(3, 1346);
		this.eLine6.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CCDB51 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		this.eLine6.Size = new System.Drawing.Size(928, 2);
		this.chk替换库存.Always = true;
		this.chk替换库存.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -1768974667 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
		this.chk替换库存.Location = new System.Drawing.Point(159, 840);
		this.chk替换库存.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x20063C6A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f);
		this.chk替换库存.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk替换库存.Size = new System.Drawing.Size(249, 28);
		this.chk替换库存.TabIndex = 574;
		this.chk替换库存.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5023438B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841);
		this.chk替换尺寸.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BD6ABB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef);
		this.chk替换尺寸.Location = new System.Drawing.Point(404, 874);
		this.chk替换尺寸.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6366F41B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3);
		this.chk替换尺寸.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk替换尺寸.Size = new System.Drawing.Size(469, 28);
		this.chk替换尺寸.TabIndex = 577;
		this.chk替换尺寸.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x103652A8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
		this.chk替换尺寸.Visible = false;
		this.eLine7.Location = new System.Drawing.Point(3, 1079);
		this.eLine7.Margin = new System.Windows.Forms.Padding(3, 5, 3, 5);
		this.eLine7.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643B7BCC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		this.eLine7.Size = new System.Drawing.Size(928, 2);
		this.ly运费.Controls.Add(this.lb运费);
		this.ly运费.Controls.Add(this.chk运费开关);
		this.ly运费.Controls.Add(this.lb模板);
		this.ly运费.Controls.Add(this.chk允许币种);
		this.ly运费.Controls.Add(this.chk排除平台);
		this.ly运费.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.ly运费.Location = new System.Drawing.Point(0, 972);
		this.ly运费.Margin = new System.Windows.Forms.Padding(0);
		this.ly运费.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x636336B5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3);
		this.ly运费.Size = new System.Drawing.Size(934, 102);
		this.ly运费.TabIndex = 582;
		this.ly运费.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63976EA8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2);
		this.lb运费.FlowBreak = false;
		this.lb运费.Location = new System.Drawing.Point(3, 3);
		this.lb运费.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C43120 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.lb运费.Padding = new System.Windows.Forms.Padding(0, 4, 0, 4);
		this.lb运费.Size = new System.Drawing.Size(71, 28);
		this.lb运费.TabIndex = 583;
		this.lb运费.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA58BE79 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1);
		this.chk运费开关.AutoWidth = true;
		this.chk运费开关.FlowBreak = false;
		this.chk运费开关.ForeColor = System.Drawing.Color.Blue;
		this.chk运费开关.LabelColor = System.Drawing.Color.Black;
		this.chk运费开关.Location = new System.Drawing.Point(80, 3);
		this.chk运费开关.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A51533C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39);
		this.chk运费开关.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk运费开关.Size = new System.Drawing.Size(463, 28);
		this.chk运费开关.TabIndex = 586;
		this.chk运费开关.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x67F98A15 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8);
		this.lb模板.Fonts = ZYing.UI.EFonts.微软雅黑Bold;
		this.lb模板.Location = new System.Drawing.Point(549, 3);
		this.lb模板.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E34B9C8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.lb模板.Padding = new System.Windows.Forms.Padding(8, 4, 0, 4);
		this.lb模板.Size = new System.Drawing.Size(79, 27);
		this.lb模板.TabIndex = 586;
		this.lb模板.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B187FAF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791);
		this.chk允许币种.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.chk允许币种.Indent = 70;
		this.chk允许币种.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x61CC1EF9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4);
		this.chk允许币种.Location = new System.Drawing.Point(80, 37);
		this.chk允许币种.Margin = new System.Windows.Forms.Padding(80, 3, 3, 3);
		this.chk允许币种.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E1D0A4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce);
		this.chk允许币种.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk允许币种.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x125B1664 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2);
		this.chk允许币种.Size = new System.Drawing.Size(850, 28);
		this.chk允许币种.TabIndex = 587;
		this.chk允许币种.Visible = false;
		this.chk排除平台.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.chk排除平台.ForeColor = System.Drawing.Color.DarkRed;
		this.chk排除平台.Indent = 70;
		this.chk排除平台.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C40D52 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.chk排除平台.LabelColor = System.Drawing.Color.DarkRed;
		this.chk排除平台.Location = new System.Drawing.Point(80, 71);
		this.chk排除平台.Margin = new System.Windows.Forms.Padding(80, 3, 3, 3);
		this.chk排除平台.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643E6CAE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		this.chk排除平台.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk排除平台.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7082EF8E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834);
		this.chk排除平台.Size = new System.Drawing.Size(850, 28);
		this.chk排除平台.TabIndex = 586;
		this.chk排除平台.Visible = false;
		this.tabLans.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.tabLans.Indent = 0;
		this.tabLans.Location = new System.Drawing.Point(79, 1233);
		this.tabLans.Margin = new System.Windows.Forms.Padding(79, 3, 3, 3);
		this.tabLans.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x4F8F3047 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		this.tabLans.Size = new System.Drawing.Size(852, 34);
		this.tabLans.TabIndex = 586;
		this.txtMlMemo.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86691BF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		this.txtMlMemo.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.txtMlMemo.FlowBreak = false;
		this.txtMlMemo.Indent = 76;
		this.txtMlMemo.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6D3B00 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.txtMlMemo.Lan = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7647CACF ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
		this.txtMlMemo.Location = new System.Drawing.Point(3, 1273);
		this.txtMlMemo.MaximumSize = new System.Drawing.Size(10000, 107);
		this.txtMlMemo.MaxLines = 5;
		this.txtMlMemo.MinimumSize = new System.Drawing.Size(0, 67);
		this.txtMlMemo.MinLines = 3;
		this.txtMlMemo.Multi = multi;
		this.txtMlMemo.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB42E7F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4);
		this.txtMlMemo.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtMlMemo.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x605D283 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2);
		this.txtMlMemo.Size = new System.Drawing.Size(461, 67);
		this.txtMlMemo.TabIndex = 588;
		this.txtMlMemo.TabStop = false;
		this.txtMlTail.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA7AA15 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
		this.txtMlTail.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.txtMlTail.Indent = 76;
		this.txtMlTail.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -1951386693 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010);
		this.txtMlTail.Lan = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC4A2B0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332);
		this.txtMlTail.Location = new System.Drawing.Point(470, 1273);
		this.txtMlTail.MaximumSize = new System.Drawing.Size(10000, 107);
		this.txtMlTail.MaxLines = 5;
		this.txtMlTail.MinimumSize = new System.Drawing.Size(0, 67);
		this.txtMlTail.MinLines = 3;
		this.txtMlTail.Multi = multi2;
		this.txtMlTail.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9FC505 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947);
		this.txtMlTail.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtMlTail.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x246DA6B4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552);
		this.txtMlTail.Size = new System.Drawing.Size(461, 67);
		this.txtMlTail.TabIndex = 589;
		this.txtMlTail.TabStop = false;
		this.chk不联动分类.Location = new System.Drawing.Point(800, 549);
		this.chk不联动分类.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D362C9A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
		this.chk不联动分类.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk不联动分类.Size = new System.Drawing.Size(119, 28);
		this.chk不联动分类.TabIndex = 591;
		this.chk不联动分类.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76455A65 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
		this.attrs.Always = true;
		this.attrs.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.attrs.Indent = 76;
		this.attrs.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B26EFE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.attrs.Location = new System.Drawing.Point(3, 700);
		this.attrs.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE4AECC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde);
		this.attrs.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.attrs.Size = new System.Drawing.Size(928, 20);
		this.attrs.TabIndex = 395;
		this.sel尺码表.AutoHeight = true;
		this.sel尺码表.Enabled = false;
		this.sel尺码表.Indent = 76;
		this.sel尺码表.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA5C5C83 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1);
		this.sel尺码表.Location = new System.Drawing.Point(3, 772);
		this.sel尺码表.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D297328 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3);
		this.sel尺码表.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.sel尺码表.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63937E40 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2);
		this.sel尺码表.RightEIcos = ZYing.IcoEnums.search;
		this.sel尺码表.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0BFC53 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.sel尺码表.Size = new System.Drawing.Size(474, 28);
		this.sel尺码表.TabIndex = 597;
		this.chk替换.Always = true;
		this.chk替换.AutoWidth = true;
		this.chk替换.Checked = true;
		this.chk替换.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D69070 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda);
		this.chk替换.ForeColor = System.Drawing.Color.Blue;
		this.chk替换.Indent = 74;
		this.chk替换.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E443B75 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		this.chk替换.LabelColor = System.Drawing.Color.Black;
		this.chk替换.Location = new System.Drawing.Point(3, 1131);
		this.chk替换.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C40CEC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.chk替换.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.chk替换.Size = new System.Drawing.Size(501, 28);
		this.chk替换.TabIndex = 601;
		this.chk替换.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F250A44 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a);
		this.selListType.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -1658075709 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.selListType.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.selListType.Indent = 76;
		this.selListType.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -448067194 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7e67c0f7adb45fc8cb0c8f442d841f1);
		this.selListType.Location = new System.Drawing.Point(470, 1089);
		this.selListType.MaxSelect = 1;
		this.selListType.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1031317C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
		this.selListType.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.selListType.Size = new System.Drawing.Size(461, 28);
		this.selListType.TabIndex = 603;
		this.sel定价模式.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.sel定价模式.Fonts = ZYing.UI.EFonts.微软雅黑Bold;
		this.sel定价模式.ForeColor = System.Drawing.Color.Blue;
		this.sel定价模式.Indent = 76;
		this.sel定价模式.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F250AEE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a);
		this.sel定价模式.LabelColor = System.Drawing.Color.Black;
		this.sel定价模式.Location = new System.Drawing.Point(3, 942);
		this.sel定价模式.MaxSelect = 1;
		this.sel定价模式.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA3B54D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
		this.sel定价模式.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.sel定价模式.Size = new System.Drawing.Size(928, 27);
		this.sel定价模式.TabIndex = 606;
		this.m产品分类.AutoHeight = true;
		this.m产品分类.Enabled = false;
		this.m产品分类.Field = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643AA18A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		this.m产品分类.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.m产品分类.FlowBreak = false;
		this.m产品分类.FlowWrap = false;
		this.m产品分类.Indent = 76;
		this.m产品分类.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6E454E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.m产品分类.Location = new System.Drawing.Point(3, 630);
		this.m产品分类.MaximumSize = new System.Drawing.Size(900, 0);
		this.m产品分类.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4244E6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835);
		this.m产品分类.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.m产品分类.Size = new System.Drawing.Size(886, 28);
		this.m产品分类.TabIndex = 8;
		base.Controls.Add(this.eLine5);
		base.Controls.Add(this.lblHelp);
		base.Controls.Add(this.gType);
		base.Controls.Add(this.gLevel);
		base.Controls.Add(this.chkGrade);
		base.Controls.Add(this.chk编号);
		base.Controls.Add(this.txtIDs);
		base.Controls.Add(this.txtTop);
		base.Controls.Add(this.selDate);
		base.Controls.Add(this.PLogin);
		base.Controls.Add(this.txtLow);
		base.Controls.Add(this.txtHigh);
		base.Controls.Add(this.PSource);
		base.Controls.Add(this.selArea);
		base.Controls.Add(this.txt采集店铺);
		base.Controls.Add(this.txtWord);
		base.Controls.Add(this.chk重量过滤);
		base.Controls.Add(this.grpSort);
		base.Controls.Add(this.eLine1);
		base.Controls.Add(this.sel站点);
		base.Controls.Add(this.lbl授权);
		base.Controls.Add(this.loader);
		base.Controls.Add(this.chk不联动分类);
		base.Controls.Add(this.grpAbbrs);
		base.Controls.Add(this.eLine2);
		base.Controls.Add(this.m产品分类);
		base.Controls.Add(this.lbl历史);
		base.Controls.Add(this.meli分类);
		base.Controls.Add(this.lbl查找meli);
		base.Controls.Add(this.attrs);
		base.Controls.Add(this.eLine3);
		base.Controls.Add(this.参数补充);
		base.Controls.Add(this.sel尺码表);
		base.Controls.Add(this.eBrand);
		base.Controls.Add(this.txt库存);
		base.Controls.Add(this.chk替换库存);
		base.Controls.Add(this.txtWeight);
		base.Controls.Add(this.txtSize);
		base.Controls.Add(this.chk替换尺寸);
		base.Controls.Add(this.sel默认尺寸);
		base.Controls.Add(this.sel定价模式);
		base.Controls.Add(this.ly运费);
		base.Controls.Add(this.eLine7);
		base.Controls.Add(this.sel保证类型);
		base.Controls.Add(this.selListType);
		base.Controls.Add(this.eLine4);
		base.Controls.Add(this.chk替换);
		base.Controls.Add(this.txtPrefix);
		base.Controls.Add(this.txtSuffix);
		base.Controls.Add(this.txt替换);
		base.Controls.Add(this.tabLans);
		base.Controls.Add(this.txtMlMemo);
		base.Controls.Add(this.txtMlTail);
		base.Controls.Add(this.eLine6);
		base.Controls.Add(this.btnSubmit);
		base.Controls.Add(this.lblOK);
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E409D62 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee);
		this.Padding = new System.Windows.Forms.Padding(0, 0, 8, 0);
		base.Size = new System.Drawing.Size(943, 1421);
		this.ly运费.ResumeLayout(false);
		this.ly运费.PerformLayout();
		base.ResumeLayout(false);
		base.PerformLayout();
	}
}
