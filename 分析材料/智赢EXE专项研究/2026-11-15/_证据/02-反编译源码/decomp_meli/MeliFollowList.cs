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
using ZYing.Interface;
using ZYing.Meli;
using ZYing.Meli.Model;
using ZYing.Static;
using ZYing.UI;

namespace ZYing.Channel;

[ToolboxItem(false)]
public class Meli跟卖列表 : CTablePanelScroll
{
	[CompilerGenerated]
	private sealed class <>c__DisplayClass10_0
	{
		public Meli跟卖列表 <>4__this;

		public ETCellOP c;

		public <>c__DisplayClass10_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}

		internal void <Tbl_OPClick>b__1(ZClient <p0>)
		{
			<>4__this.DoFresh(<>4__this, EventArgs.Empty);
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass10_1
	{
		public ZClient zc;

		public string key;

		public <>c__DisplayClass10_0 CS$<>8__locals1;

		public <>c__DisplayClass10_1()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}

		internal void <Tbl_OPClick>b__2(ZClient <p0>)
		{
			if (zc.Result > 0)
			{
				CS$<>8__locals1.c.Table.Remove(key);
			}
		}
	}

	public const string Key = "catalogs";

	public const string Caption = "目录列表";

	private IContainer components;

	private ELoader lblTotal;

	private ETurn PTurn;

	private ECombo PAcc;

	private EDropLinkage PKind;

	private ESelectDateRange PDate;

	private ESearch PWord;

	private ETabRect abbrTab;

	private ELine eLine1;

	private EFilterDrop pageFilter;

	public static Meli跟卖列表 Instance => Global.Create(typeof(Meli跟卖列表)) as Meli跟卖列表;

	public Meli跟卖列表()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		right.Icon = Icos.Get(IcoEnums.article);
		tbl.RowHeight = 58;
		tbl.Columns.AddRange(new ETColumn(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63979636 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB5FB33 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4), 26, 0, ETCellStyle.Pick), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5110B091 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3A2316 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273), 58, 58, ETCellStyle.Image), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE3D252 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB31533 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4), 70, int.MaxValue, ETCellStyle.Multi, ETheme.Verdana), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224B6734 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7086606A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834), 60, 80, ETCellStyle.Multi), new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E4AC77 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68D0419A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6), 60, 200, ETheme.微软雅黑Bold), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1032AB54 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA8032C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b), 60, 250, ETCellStyle.Multi), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5020DB85 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CC3789 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206), 60, 200, ETCellStyle.Multi), new ETColumn(7, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2003769E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51108137 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7), 60, 100, ETCellStyle.Multi), new ETColumn(13, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -1921104813 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468F86AE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6), 100, 0, ETCellStyle.Multi, ETheme.Verdana), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2002F650 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B264CC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53), 50, 50, ETCellStyle.OP));
		tbl.RowCreating = tbl_RowCreating;
		tbl.RowSetting = tbl_RowSetting;
		tbl.CellOPClick = Tbl_OPClick;
		ETable eTable = tbl;
		eTable.OrderChanged = (Action<ETColumn>)Delegate.Combine(eTable.OrderChanged, (Action<ETColumn>)((ETColumn cl) =>
		{
			DoSelect(null, EventArgs.Empty);
		}));
		pageFilter.Items = new EItemCollection(new EItem(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B227F7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A8BFA6A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21)));
		PAcc.Items = Enumer.Meli.CBTAcc;
		PAcc.SelectedIndex = 0;
		PAcc.SelectedChanged += DoSelect;
		abbrTab.Items = Enumer.Meli.Abbrs.Copy();
		abbrTab.Items.RemoveAt(0);
		abbrTab.SelectedItem = abbrTab.Items.First;
		abbrTab.SelectedChanged += AbbrTab_SelectedChanged;
		PKind.DropClosed += DoStat;
		PKind.Kinder = KindMeli.Get(abbrTab.SelectedItem.Key);
		PWord.OK += DoStat;
		PWord.Void += DoVoid;
		PDate.DateChanged += DoSelect;
		PDate.Start.PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E41B01C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee);
		PDate.End.PlaceHolder = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7646EA7F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
		PTurn.PageChanged += DoFresh;
	}

	private void AbbrTab_SelectedChanged(object sender, EItemArgs e)
	{
		PKind.Kinder = KindMeli.Get(abbrTab.SelectedItem.Key);
		DoSelect(sender, e);
	}

	public override void Load()
	{
		DoStat(this, EventArgs.Empty);
	}

	protected override EControl OnCreateRight(ISource src, string name)
	{
		return new Meli跟卖详情(src, name);
	}

	protected override void tbl_RowCreating(ETRow tr)
	{
		if (tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BDC1E3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef)] is ETCellMulti eTCellMulti)
		{
			eTCellMulti.SubFont = ETheme.Verdana;
		}
		if (tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -29950230 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841)] is ETCellMulti eTCellMulti2)
		{
			eTCellMulti2.SubFont = ETheme.Verdana;
		}
		((ETCellOP)tr.LastCell).AddRange(new UIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6397FF32 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2)));
	}

	protected override void tbl_RowSetting(ETRow tr, Record re)
	{
		tr.Key = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA7DACB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97)) + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FA14E3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5) + re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B25836 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2));
		if (re.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA80324 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b)))
		{
			string abbr = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B4281F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552));
			int num = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CD01FD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206));
			tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9EFA63 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947)] = KindMeli.Get(abbr).GetFullName(num);
			tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B292A4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53)] = KindMeli.Get(abbr).GetFullSub(num);
		}
		ETCell eTCell = tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x70878DF2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834)];
		string name = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527B5B5B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1));
		eTCell.Icon = Icos.Get(name);
		string value = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA5991B5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5de0933087984998b9662cb6a4be36f1));
		eTCell = tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA810DC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b)];
		eTCell.Text = "";
		if (!value.IsNullOrEmpty())
		{
			eTCell.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA02C9D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
			eTCell.Color = Color.Blue;
			if (re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A69C110 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535)))
			{
				bool flag = re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5020D815 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841));
				eTCell.Text = (flag ? kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C9B3F43 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947) : kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B40E82 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2));
				eTCell.Color = (flag ? Color.DarkGreen : Color.Chocolate);
			}
		}
	}

	private void Tbl_OPClick(ETCellOP c, UIcon ico)
	{
		<>c__DisplayClass10_0 obj = new <>c__DisplayClass10_0();
		obj.<>4__this = this;
		obj.c = c;
		<>c__DisplayClass10_1 CS$<>8__locals18 = new <>c__DisplayClass10_1();
		CS$<>8__locals18.CS$<>8__locals1 = obj;
		if (ico.Name == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA77C37 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97))
		{
			if (!EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78E1293B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c)) || !EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E43F211 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689)))
			{
				return;
			}
			CS$<>8__locals18.zc = ZClient.Create();
			CS$<>8__locals18.zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E430436 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee);
			CS$<>8__locals18.zc.Failed = (Exception ex) =>
			{
				EDialog.Show(ex);
			};
			CS$<>8__locals18.zc.Finish = delegate
			{
				CS$<>8__locals18.CS$<>8__locals1.<>4__this.DoFresh(CS$<>8__locals18.CS$<>8__locals1.<>4__this, EventArgs.Empty);
			};
			NameValueCollection nameValueCollection = new NameValueCollection();
			CS$<>8__locals18.key = CS$<>8__locals18.CS$<>8__locals1.c.Row.Key;
			int num = CS$<>8__locals18.key.IndexOf('-');
			nameValueCollection[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x32244B4C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594)] = CS$<>8__locals18.key.Substring(0, num);
			nameValueCollection[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x70869896 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834)] = CS$<>8__locals18.key.Substring(num + 1);
			CS$<>8__locals18.zc.Finish = delegate
			{
				if (CS$<>8__locals18.zc.Result > 0)
				{
					CS$<>8__locals18.CS$<>8__locals1.c.Table.Remove(CS$<>8__locals18.key);
				}
			};
			CS$<>8__locals18.zc.PostAsync(nameValueCollection);
		}
		else
		{
			tbl.CurrentRow = CS$<>8__locals18.CS$<>8__locals1.c.Row;
		}
	}

	private Para DoPara(bool select)
	{
		Para para = new Para
		{
			order = tbl.OrderString,
			localid = Valid.ToInt(PKind.Value),
			aid = Valid.ToInt(PAcc.Value),
			start = PDate.Start.Value,
			end = PDate.End.Value,
			word = PWord.Text,
			pagesize = Valid.ToInt(pageFilter.Value),
			select = select,
			page = PTurn.Page,
			abbr = abbrTab.Value
		};
		para.id = Valid.ToInt(para.word);
		if (para.id > 0)
		{
			para.word = null;
		}
		return para;
	}

	private void DoVoid(object sender, EventArgs e)
	{
		PKind.Reset();
		PAcc.Reset();
		PWord.Text = "";
		if (tbl.OrderColumn != null)
		{
			tbl.OrderColumn.Order = OrderBy.None;
		}
		DoStat(sender, e);
	}

	private void DoStat(object sender, EventArgs e)
	{
		DoCommand(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x665555A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8), DoPara(select: true));
	}

	private void DoSelect(object sender, EventArgs e)
	{
		DoCommand(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x105317AA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf), DoPara(select: true));
	}

	private void DoFresh(object sender, EventArgs e)
	{
		Para para = DoPara(select: true);
		para.maxcount = PTurn.MaxCount;
		para.page = PTurn.Page;
		DoCommand(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D37D21C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b), para);
	}

	private async void DoCommand(string cmd, Para p)
	{
		if (p.aid > 0)
		{
			Record record = await LoadItems(await MeliApiBuilder.Create(p.aid), PTurn.Page);
			tbl.DataSource = record;
			PTurn.Read(record);
			ELoader eLoader = lblTotal;
			DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(6, 1);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7A3A3DF8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273));
			defaultInterpolatedStringHandler.AppendFormatted(record.MaxCount);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408A348C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44));
			eLoader.Text = defaultInterpolatedStringHandler.ToStringAndClear();
		}
	}

	private async Task<Record> LoadItems(MeliApi api, int page)
	{
		_ = 1;
		try
		{
			lblTotal.Loading = true;
			MeliItemQuery meliItemQuery = new MeliItemQuery
			{
				limit = 50,
				q = PWord.Text
			};
			meliItemQuery.Page = page;
			SearchSiteItemsResult searchSiteItemsResult = await api.SearchItemsBySeller(meliItemQuery);
			Record re = new Record
			{
				Page = page,
				PageSize = 50,
				MaxCount = searchSiteItemsResult.Total
			};
			foreach (MeliItemInfo result in searchSiteItemsResult.results)
			{
				re.AddRow();
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63971110 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2)).Set(api.Aid);
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643BA27A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848)).Set(result.id);
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A558D10 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39)).Set(result.title);
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D11352 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda)).Set(result.site_id);
				string value = ((!result.category_id.IsNullOrEmpty() && result.category_id.Length >= 3) ? result.category_id.Substring(3) : "");
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D358266 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b)).Set(value);
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -338024309 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9)).Set(result.domain_id);
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D6B358A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9)).Set(result.currency_id);
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB13CA1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018)).Set(result.price);
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -950135924 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8)).Set(result.available_quantity);
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACE56F6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057)).Set(result.thumbnail);
				re.Columns.AddBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x50170386 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e)).Set(result.catalog_listing);
				re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC49D0E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332)).Set(result.catalog_product_id);
				if (result.catalog_listing)
				{
					PriceToWinResult priceToWinResult = await api.PriceToWin(result.id);
					re.Columns.AddBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA02CCB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97)).Set(priceToWinResult.HasWinner);
					re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA3924A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b)).Set(priceToWinResult.WinnerItemId);
				}
			}
			return re;
		}
		finally
		{
			lblTotal.Loading = false;
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
		this.PTurn = new ZYing.UI.ETurn();
		this.lblTotal = new ZYing.UI.ELoader();
		this.PAcc = new ZYing.UI.ECombo();
		this.PKind = new ZYing.UI.EDropLinkage();
		this.PDate = new ZYing.UI.ESelectDateRange();
		this.PWord = new ZYing.UI.ESearch();
		this.abbrTab = new ZYing.UI.ETabRect();
		this.eLine1 = new ZYing.UI.ELine();
		this.pageFilter = new ZYing.UI.EFilterDrop();
		base.main.SuspendLayout();
		base.tail.SuspendLayout();
		base.head.SuspendLayout();
		base.SuspendLayout();
		base.right.Location = new System.Drawing.Point(631, 0);
		base.right.MaximumSize = new System.Drawing.Size(1200, 500000);
		base.right.MinimumSize = new System.Drawing.Size(550, 0);
		base.right.Size = new System.Drawing.Size(550, 667);
		base.right.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B7EE45 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e);
		base.main.Size = new System.Drawing.Size(631, 667);
		base.tail.Controls.Add(this.PTurn);
		base.tail.Controls.Add(this.lblTotal);
		base.tail.Location = new System.Drawing.Point(7, 595);
		base.tail.Size = new System.Drawing.Size(616, 42);
		base.tbl.Location = new System.Drawing.Point(10, 111);
		base.tbl.Margin = new System.Windows.Forms.Padding(3, 0, 3, 3);
		base.tbl.MaxHeight = 504;
		base.tbl.Size = new System.Drawing.Size(610, 481);
		base.tbl.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x32217564 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		base.head.Controls.Add(this.PAcc);
		base.head.Controls.Add(this.PKind);
		base.head.Controls.Add(this.PDate);
		base.head.Controls.Add(this.PWord);
		base.head.Controls.Add(this.eLine1);
		base.head.Controls.Add(this.abbrTab);
		base.head.Controls.Add(this.pageFilter);
		base.head.Size = new System.Drawing.Size(616, 105);
		this.PTurn.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.PTurn.FlowBreak = false;
		this.PTurn.Location = new System.Drawing.Point(3, 5);
		this.PTurn.Margin = new System.Windows.Forms.Padding(3, 5, 3, 10);
		this.PTurn.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224B0752 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2);
		this.PTurn.Padding = new System.Windows.Forms.Padding(4);
		this.PTurn.Size = new System.Drawing.Size(441, 27);
		this.PTurn.TabIndex = 10;
		this.PTurn.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10343244 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d);
		this.lblTotal.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.lblTotal.ForeColor = System.Drawing.Color.Black;
		this.lblTotal.Location = new System.Drawing.Point(450, 3);
		this.lblTotal.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2002F76C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f);
		this.lblTotal.Padding = new System.Windows.Forms.Padding(5, 4, 2, 4);
		this.lblTotal.Size = new System.Drawing.Size(162, 28);
		this.lblTotal.TabIndex = 10;
		this.lblTotal.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3224EDDA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		this.lblTotal.TextAlign = System.Windows.Forms.HorizontalAlignment.Right;
		this.PAcc.AutoHeight = true;
		this.PAcc.FlowBreak = false;
		this.PAcc.FlowWrap = false;
		this.PAcc.IsExact = true;
		this.PAcc.Location = new System.Drawing.Point(3, 3);
		this.PAcc.MaximumSize = new System.Drawing.Size(280, 28);
		this.PAcc.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DC4AA26 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332);
		this.PAcc.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PAcc.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x8611ADD ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		this.PAcc.Size = new System.Drawing.Size(222, 28);
		this.PAcc.TabIndex = 247;
		this.PKind.AutoHeight = true;
		this.PKind.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.PKind.Location = new System.Drawing.Point(231, 3);
		this.PKind.MaximumSize = new System.Drawing.Size(450, 27);
		this.PKind.MinimumSize = new System.Drawing.Size(150, 0);
		this.PKind.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86693E7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		this.PKind.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PKind.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -936188857 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e);
		this.PKind.Size = new System.Drawing.Size(381, 27);
		this.PKind.TabIndex = 249;
		this.PDate.FlowBreak = false;
		this.PDate.Location = new System.Drawing.Point(3, 37);
		this.PDate.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63672AF9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3);
		this.PDate.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PDate.Size = new System.Drawing.Size(222, 28);
		this.PDate.TabIndex = 248;
		this.PWord.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.PWord.FlowBreak = false;
		this.PWord.Location = new System.Drawing.Point(231, 37);
		this.PWord.MaxLength = 2000;
		this.PWord.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-813636989) ^ 0x3A7028BE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.PWord.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PWord.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B7F722 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53);
		this.PWord.Resetable = true;
		this.PWord.Size = new System.Drawing.Size(187, 28);
		this.PWord.TabIndex = 245;
		this.abbrTab.AutoSize = true;
		this.abbrTab.BottomLine = false;
		this.abbrTab.CellPadding = new System.Windows.Forms.Padding(10, 4, 10, 4);
		this.abbrTab.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.abbrTab.FlowBreak = false;
		this.abbrTab.Indent = 0;
		this.abbrTab.Location = new System.Drawing.Point(3, 73);
		this.abbrTab.Margin = new System.Windows.Forms.Padding(3, 5, 3, 0);
		this.abbrTab.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0AD49F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.abbrTab.Padding = new System.Windows.Forms.Padding(0, 0, 0, 3);
		this.abbrTab.Size = new System.Drawing.Size(587, 32);
		this.abbrTab.TabIndex = 159;
		this.abbrTab.Visible = false;
		this.eLine1.Location = new System.Drawing.Point(424, 37);
		this.eLine1.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1051F454 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf);
		this.eLine1.Size = new System.Drawing.Size(188, 2);
		this.pageFilter.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.pageFilter.Location = new System.Drawing.Point(596, 68);
		this.pageFilter.Margin = new System.Windows.Forms.Padding(3, 0, 0, 1);
		this.pageFilter.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F47372C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835);
		this.pageFilter.Padding = new System.Windows.Forms.Padding(0, 3, 5, 3);
		this.pageFilter.Size = new System.Drawing.Size(19, 26);
		this.pageFilter.TabIndex = 12;
		this.pageFilter.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E423B61 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_84e246401c064d72bf30c1a285ffd689);
		this.pageFilter.VLine = false;
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C13D9C6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1);
		base.Size = new System.Drawing.Size(1181, 667);
		base.main.ResumeLayout(false);
		base.tail.ResumeLayout(false);
		base.head.ResumeLayout(false);
		base.head.PerformLayout();
		base.ResumeLayout(false);
	}
}
