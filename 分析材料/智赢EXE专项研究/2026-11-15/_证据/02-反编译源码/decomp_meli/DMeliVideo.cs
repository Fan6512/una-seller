using System;
using System.Collections.Generic;
using System.Collections.Specialized;
using System.ComponentModel;
using System.Drawing;
using System.IO;
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

namespace ZYing.Dialog;

[ToolboxItem(false)]
public class DMeli商品视频 : EDialogLayer
{
	[CompilerGenerated]
	private sealed class <>c__DisplayClass14_0
	{
		public MeliClipsRequest req;

		public <>c__DisplayClass14_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.fmiGD6eDN();
			base..ctor();
		}

		internal void <Btn上传_Click>b__0(EItem x)
		{
			req.sites.Add(new MarketplaceInfo
			{
				site_id = x.Key,
				logistic_type = x.Tag?.ToString()
			});
		}
	}

	private int _aid;

	private int _groupid;

	private string _cbtid;

	private string _itemid;

	private IContainer components;

	private EButton btnCancel;

	private ETable table;

	private ELoader loader;

	private ESearch PWord;

	private EButton btn同步;

	private ELayer header;

	private ELabel lbcbtId;

	private ELine eLine1;

	private ELabel lbVideo;

	private EGroup grpSites;

	private EButton btn上传;

	private ECombo PAcc;

	public static DMeli商品视频 Instance => Global.Create(typeof(DMeli商品视频)) as DMeli商品视频;

	public DMeli商品视频()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		table.RowHeight = 55;
		table.Columns.AddRange(new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408BBA7E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B3DE5B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552), 0, 0, hidden: true), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CC3979 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AC3A88 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611), 0, 0, hidden: true), new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A54A66A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68AC3A58 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_49435e5d67ab4ac985caeca22a8c8611), 50, 0, ETCellStyle.Multi), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DC9345A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C15B562 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1), 50, 200, ETCellStyle.Multi), new ETColumn(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D77EB8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x468FCA2A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6), 50, 100, ETCellStyle.Multi), new ETColumn(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1634166D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C15BAD8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1), 80, int.MaxValue, ETCellStyle.Multi), new ETColumn(9, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E3143AC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B304CA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2), 80, 0, ETCellStyle.Multi), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D642A0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0F7793 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd), 50, 70, ETCellStyle.OP));
		table.Adjust();
		table.RowCreating = tbl_RowCreating;
		table.RowSetting = tbl_RowSetting;
		table.CellOPClick = tbl_OPClick;
		PAcc.Items = Enumer.Meli.AllCBTGroupAccs;
		PWord.OK += async delegate
		{
			await doCommand(doPara());
		};
		PWord.Void += async delegate
		{
			PWord.Reset();
			await doCommand(doPara());
		};
		btnCancel.Click += delegate
		{
			Hide();
		};
		lbVideo.RightClick += delegate
		{
			OpenFileDialog openFileDialog = new OpenFileDialog
			{
				AddExtension = true,
				CheckFileExists = true,
				CheckPathExists = true,
				Multiselect = false,
				Filter = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D6E2A94 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9)
			};
			if (openFileDialog.ShowDialog() == DialogResult.OK)
			{
				lbVideo.Text = openFileDialog.FileName;
				btn上传.Enabled = !lbVideo.Text.IsNullOrEmpty() && File.Exists(lbVideo.Text);
			}
		};
		btn上传.Click += Btn上传_Click;
		btn同步.Click += Btn同步_Click;
	}

	private void tbl_RowCreating(ETRow tr)
	{
		((ETCellOP)tr.LastCell).AddRange(new UIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CD04B7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206)));
	}

	private void tbl_RowSetting(ETRow tr, Record re)
	{
		string key = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45D65558 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda));
		EItem eItem = MeliClips.Status.FindKey(key);
		ETCellMulti obj = tr.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37B0A8B1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e)] as ETCellMulti;
		obj.Text = eItem?.Text;
		obj.Color = eItem?.Color ?? Color.Empty;
		tr[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7087154C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834)] = re.GetDateText(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B1D8563 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f40a764b40734312857dfc48304a3791), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F218B2A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a));
	}

	private async void tbl_OPClick(ETCellOP c, UIcon ico)
	{
		if (ico.Name == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -846747451 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_aa25f6befa7a4ce1b0bccf0e5571b7d2))
		{
			if (!EDialog.Confirm(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x502E1B33 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841)))
			{
				return;
			}
			ZClient zc = ZClient.Create(ZClient.美客多);
			zc.Loader = loader;
			zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78ED2933 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c);
			string uuid = c.Row[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x50153F3E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e)];
			NameValueCollection nv = new NameValueCollection
			{
				[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-813636989) ^ 0x4A45E20B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_897ed775279c421889b43f6742093273)] = _aid.ToString(),
				[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x63608AB1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3)] = _cbtid,
				[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E351D46 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6)] = uuid
			};
			await zc.PostAsync(nv);
			zc.Throw();
			ETRow[] array = c.Table.Rows.ToArray();
			foreach (ETRow eTRow in array)
			{
				if (eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA21055 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97)] == uuid)
				{
					c.Table.Rows.Remove(eTRow);
				}
			}
		}
		else
		{
			table.CurrentRow = c.Row;
		}
	}

	public async Task Show(EControl parent, int aid, string cbtid, string itemid)
	{
		_aid = aid;
		_cbtid = cbtid;
		_itemid = itemid;
		Reset();
		lbcbtId.Text = _cbtid;
		PWord.Text = _cbtid;
		EItem eItem = Enumer.Meli.AllCBTGroupAccs.FindKey(aid);
		PAcc.Id = eItem?.Group ?? 0;
		ZClient zc = ZClient.Create(ZClient.美客多);
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x322D7FD0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		zc.Loader = loader;
		NameValueCollection nv = new NameValueCollection
		{
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D3521D0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b)] = _aid.ToString(),
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA008BB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97)] = _cbtid
		};
		await zc.PostAsync(nv);
		zc.Throw();
		Record record = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3224A56E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594)];
		table.DataSource = record;
		if (record.Count == 0)
		{
			Record record2 = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46864F0A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_16183a7e5f414a628da6e231cce406e6)];
			EItemCollection eItemCollection = new EItemCollection();
			while (record2.Read())
			{
				record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E456B0A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee));
				string text = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x51152BA3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7));
				string tag = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408F45CE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44));
				eItemCollection.Add(new EItem(text, Enumer.Meli.GetCNName(text) ?? "")
				{
					Tag = tag
				});
			}
			grpSites.Items = eItemCollection;
		}
		base.Show(parent);
	}

	public override void Reset()
	{
		lbVideo.Text = "";
		btn上传.Enabled = false;
		grpSites.Items.Clear();
		PWord.Reset();
		table.Rows.Clear();
	}

	private Para doPara()
	{
		return new Para
		{
			pid = (PAcc.SelectedItem?.Group ?? 0),
			word = PWord.Text
		};
	}

	private async Task doCommand(Para p)
	{
		ZClient zc = ZClient.Create(ZClient.美客多);
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1705051699 ^ -226017983 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6);
		zc.Loader = loader;
		await zc.PostAsync(p);
		zc.Throw();
		table.DataSource = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x50101B82 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e)];
	}

	private async void Btn上传_Click(object sender, EventArgs e)
	{
		string path = lbVideo.Text;
		if (!File.Exists(path))
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86775C9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788) + path + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FE455 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2));
		}
		try
		{
			<>c__DisplayClass14_0 CS$<>8__locals3 = new <>c__DisplayClass14_0();
			btn上传.Enabled = false;
			loader.Loading = true;
			byte[] bytes = File.ReadAllBytes(path);
			MeliApi api = await MeliApiBuilder.Create(_aid);
			CS$<>8__locals3.req = new MeliClipsRequest
			{
				cbt_item_id = _cbtid,
				file_name = Path.GetFileName(path),
				content = bytes,
				sites = new List<MarketplaceInfo>()
			};
			if (grpSites.SelectedItems.Count > 0)
			{
				grpSites.SelectedItems.ForEach((EItem x) =>
				{
					CS$<>8__locals3.req.sites.Add(new MarketplaceInfo
					{
						site_id = x.Key,
						logistic_type = x.Tag?.ToString()
					});
				});
			}
			if ((await api.UploadClips(CS$<>8__locals3.req)).clip_uuid.IsNullOrEmpty())
			{
				throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A5C07AC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7d81c898f1574d79934ac53b0b17ff39));
			}
			MeliClipsInfo meliClipsInfo = await api.GetClips(_cbtid);
			Record re = new Record();
			foreach (Clip clip in meliClipsInfo.clips)
			{
				foreach (Metadata item in clip.metadata)
				{
					if (!(item.status == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C95F313 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f84ef4f5afe4432b0ec87538d4c5947)))
					{
						re.AddRow();
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C1D143C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3f6eb617a9b34fbc8fc3e1f3386786a1)).Set(item.item_id);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x45DFDE7E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda)).Set(meliClipsInfo.parent_item_id);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2E4DC980 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1fdbfc75f2b249b6ba26c2581cf40fee)).Set(clip.clip_uuid);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25AA1B66 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010)).Set(item.user_id);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EA19BA8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b)).Set(item.site_id);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22BAC374 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2)).Set(api.GroupId);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25AA1BC0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010)).Set(_aid);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1055B56A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf)).Set(item.status);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FE55B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5a6cf4329c044827b99d2026250f1eb2)).Set(path);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67BAE5F1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552)).Set(Environment.MachineName);
						re.Columns.AddDate(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408281FE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44)).Set(DateTime.Now);
						re.Columns.AddDate(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5019C1C0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e)).Set(DateTime.Now);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x78ED2879 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5739b7eb780f40948bd32d85fcc3134c)).Set(Member.UserID);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x66C96E8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8)).Set(Member.LoginID);
						re.Columns.Add(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FBE1671 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4)).Set(Member.Name);
					}
				}
			}
			if (re.Count == 0)
			{
				throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7DCC0BA0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_21698db7fb084c70bf866ff74b40b332));
			}
			ZClient zc = ZClient.Create(ZClient.美客多);
			zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x502E1AD3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841);
			await zc.PostAsync(re);
			zc.Throw();
			table.DataSource = re.Abbr();
		}
		finally
		{
			btn上传.Enabled = true;
			loader.Loading = false;
		}
	}

	private async void Btn同步_Click(object sender, EventArgs e)
	{
		ZClient zc = ZClient.Create(ZClient.美客多);
		zc.Button = btn同步;
		zc.Loader = loader;
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x708FB5B4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834);
		NameValueCollection nv = new NameValueCollection
		{
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE427F2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde)] = _aid.ToString(),
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA084115 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd)] = _cbtid
		};
		await zc.PostAsync(nv);
		zc.Throw();
		Record record = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2FB7CC59 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9ac8d2aa63f7410c8b6321d67d2446a4)];
		if (record.Count > 0)
		{
			table.DataSource = record;
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
		this.btnCancel = new ZYing.UI.EButton();
		this.loader = new ZYing.UI.ELoader();
		this.table = new ZYing.UI.ETable();
		this.PWord = new ZYing.UI.ESearch();
		this.btn同步 = new ZYing.UI.EButton();
		this.header = new ZYing.UI.ELayer();
		this.lbcbtId = new ZYing.UI.ELabel();
		this.lbVideo = new ZYing.UI.ELabel();
		this.grpSites = new ZYing.UI.EGroup();
		this.btn上传 = new ZYing.UI.EButton();
		this.eLine1 = new ZYing.UI.ELine();
		this.PAcc = new ZYing.UI.ECombo();
		this.header.SuspendLayout();
		base.SuspendLayout();
		this.btnCancel.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.btnCancel.ForeColor = System.Drawing.Color.DarkRed;
		this.btnCancel.Location = new System.Drawing.Point(862, 570);
		this.btnCancel.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C0C1B6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.btnCancel.Size = new System.Drawing.Size(44, 25);
		this.btnCancel.TabIndex = 283;
		this.btnCancel.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2EA7667F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_91b545b490134d4688cb24bdb1290c97);
		this.loader.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.loader.Location = new System.Drawing.Point(502, 123);
		this.loader.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11643C1F ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca);
		this.loader.Size = new System.Drawing.Size(404, 28);
		this.loader.TabIndex = 541;
		this.table.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.table.Location = new System.Drawing.Point(15, 157);
		this.table.MaxHeight = 500;
		this.table.MinHeight = 300;
		this.table.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BD5279 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef);
		this.table.Size = new System.Drawing.Size(891, 407);
		this.table.TabIndex = 540;
		this.table.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F20EE36 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a);
		this.PWord.FlowBreak = false;
		this.PWord.Location = new System.Drawing.Point(255, 123);
		this.PWord.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2002F740 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4d35c246b8f7482f888c4f7579e7a56f);
		this.PWord.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PWord.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A8F5350 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21);
		this.PWord.Resetable = true;
		this.PWord.Size = new System.Drawing.Size(241, 28);
		this.PWord.TabIndex = 543;
		this.btn同步.FlowBreak = false;
		this.btn同步.ForeColor = System.Drawing.Color.DarkBlue;
		this.btn同步.Location = new System.Drawing.Point(15, 570);
		this.btn同步.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A2B4D0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010);
		this.btn同步.Size = new System.Drawing.Size(72, 25);
		this.btn同步.TabIndex = 544;
		this.btn同步.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E31433A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6);
		this.header.Controls.Add(this.lbcbtId);
		this.header.Controls.Add(this.lbVideo);
		this.header.Controls.Add(this.grpSites);
		this.header.Controls.Add(this.btn上传);
		this.header.Controls.Add(this.eLine1);
		this.header.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.header.Location = new System.Drawing.Point(15, 41);
		this.header.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BC5D51 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef);
		this.header.Size = new System.Drawing.Size(891, 76);
		this.header.TabIndex = 545;
		this.header.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x50108A18 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e);
		this.lbcbtId.AutoSize = false;
		this.lbcbtId.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.lbcbtId.FlowBreak = false;
		this.lbcbtId.Indent = 80;
		this.lbcbtId.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6FA3C0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.lbcbtId.Location = new System.Drawing.Point(3, 3);
		this.lbcbtId.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F20F198 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a);
		this.lbcbtId.Size = new System.Drawing.Size(439, 28);
		this.lbcbtId.TabIndex = 546;
		this.lbVideo.AutoSize = false;
		this.lbVideo.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.lbVideo.Indent = 40;
		this.lbVideo.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D34B500 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
		this.lbVideo.Location = new System.Drawing.Point(448, 3);
		this.lbVideo.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D34B57A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
		this.lbVideo.RightEIcos = ZYing.IcoEnums.dir;
		this.lbVideo.RightLabel = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408A2ECA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44);
		this.lbVideo.Size = new System.Drawing.Size(439, 28);
		this.lbVideo.TabIndex = 547;
		this.grpSites.FlowBreak = false;
		this.grpSites.Indent = 80;
		this.grpSites.Label = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACF12AE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057);
		this.grpSites.Location = new System.Drawing.Point(3, 37);
		this.grpSites.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xACF12B8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cbebca54b21b49edacbbac997877e057);
		this.grpSites.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.grpSites.ScrollFromDisplayTop = false;
		this.grpSites.Size = new System.Drawing.Size(439, 28);
		this.grpSites.TabIndex = 561;
		this.btn上传.Enabled = false;
		this.btn上传.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.btn上传.ForeColor = System.Drawing.Color.DarkBlue;
		this.btn上传.Location = new System.Drawing.Point(843, 37);
		this.btn上传.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5027ED87 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9b67005515114390b5422449ba562841);
		this.btn上传.Size = new System.Drawing.Size(44, 25);
		this.btn上传.TabIndex = 546;
		this.btn上传.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61B2A9A2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_dc096d9b09ba4c64b64e49fd44136d53);
		this.eLine1.Location = new System.Drawing.Point(3, 71);
		this.eLine1.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-2066591320) ^ 0x164201DE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9);
		this.eLine1.Size = new System.Drawing.Size(884, 2);
		this.PAcc.AutoHeight = true;
		this.PAcc.FlowBreak = false;
		this.PAcc.IsExact = true;
		this.PAcc.Location = new System.Drawing.Point(15, 123);
		this.PAcc.MaxHeight = 432;
		this.PAcc.MaximumSize = new System.Drawing.Size(280, 28);
		this.PAcc.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BD5327 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef);
		this.PAcc.NeedText = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86774D7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0ca37a00f82e496382eb1ea8d1a97788);
		this.PAcc.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PAcc.Size = new System.Drawing.Size(234, 28);
		this.PAcc.TabIndex = 546;
		base.Caption = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1050712E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_071e57290f02448f83de36be0cced6cf);
		base.ClientSize = new System.Drawing.Size(921, 613);
		base.Controls.Add(this.header);
		base.Controls.Add(this.PAcc);
		base.Controls.Add(this.PWord);
		base.Controls.Add(this.loader);
		base.Controls.Add(this.table);
		base.Controls.Add(this.btn同步);
		base.Controls.Add(this.btnCancel);
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D6E2AEE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b7da036b195c4aa882c63211c04305f9);
		base.Padding = new System.Windows.Forms.Padding(12, 38, 12, 15);
		this.header.ResumeLayout(false);
		this.header.PerformLayout();
		base.ResumeLayout(false);
		base.PerformLayout();
	}
}
