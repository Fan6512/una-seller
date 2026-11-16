using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Drawing;
using System.Drawing.Drawing2D;
using System.Windows.Forms;
using Enumer;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Auto;
using ZYing.UI;

namespace ZYing.Channel;

[ToolboxItem(false)]
public class Meli状态同步 : EContainer
{
	public const string Key = "MeliStatus";

	public const string Caption = "状态同步";

	private IContainer components;

	private EListFilter lstAcc;

	private ELabel lblAll;

	private EControlCols body;

	private ETable tbl;

	private ETextArea txtResult;

	private ELayer top;

	private EButton btnStart;

	private ELabel lblDate;

	private ELayer left;

	private EContainer main;

	private ELabel lblNote;

	private ESelectDate PStartDate;

	private ESelectDate PEndDate;

	public static Meli状态同步 Instance => Global.Create(typeof(Meli状态同步)) as Meli状态同步;

	public Meli状态同步()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.fmiGD6eDN();
		base..ctor();
		InitializeComponent();
		lstAcc.Items = Enumer.Meli.CBTAcc;
		lblAll.Click += lblAll_Click;
		PStartDate.Value = DateTime.Now.AddDays(-30.0).ToString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB659B1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018));
		tbl.Padding = new Padding(6, 5, 6, 5);
		tbl.Columns.AddRange(new ETColumn(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE427F2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67B37969 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7395fb372c534a0ea896e6536cbd6552), 40, 0, ETheme.Verdana), new ETColumn(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5011F20E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8d6acbaea9c541d4bd78d53a28d9723e), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0E3A99 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd), 60, 0), new ETColumn(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1035EEB6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -1458133444 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6), 80, int.MaxValue), new ETColumn(6, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x116553D1 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x16349F27 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d4a96a48d2ae44bd991c24d94f5ec2e8), 60, 0), new ETColumn(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BCD499 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x103406EC ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d), 40, 0, ETheme.Verdana));
		tbl.RowChanged = Tbl_RowChanged;
		btnStart.Click += delegate
		{
			if (ZSyncMeli.IsBusy)
			{
				Stop();
			}
			else
			{
				Start();
			}
		};
	}

	private void lblAll_Click(object sender, EventArgs e)
	{
		lstAcc.SelectAll();
	}

	private void Tbl_RowChanged(ETRow tr, ETRowEventArgs e)
	{
		if (tr != null)
		{
			string text = tr.Tag as string;
			txtResult.ForeColor = tr.Color;
			txtResult.Text = text ?? "";
		}
	}

	public void Stop()
	{
		ZSyncMeli.Stop();
		StartFinish();
	}

	public void Start()
	{
		try
		{
			if (ZSyncMeli.IsBusy)
			{
				return;
			}
			List<EItem> selectedItems = lstAcc.SelectedItems;
			if (selectedItems.Count == 0)
			{
				throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x527ED539 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_7adfe4eb09024f8ca7020f4c0180dad1));
			}
			任务.Instance.Enable(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE38094 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde));
			btnStart.ForeColor = Color.Red;
			btnStart.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1163E66B ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_cd794fcf8e814b138c05450caff434ca);
			tbl.Rows.Clear();
			txtResult.Text = "";
			lblDate.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D328686 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
			ZSyncMeli.Output = (string str) =>
			{
				lblDate.Text = str;
			};
			ZSyncMeli.Error = (string str) =>
			{
				ETRow lastRow = tbl.LastRow;
				if (lastRow != null)
				{
					lastRow.Tag = str;
					lastRow.Color = Color.Red;
					if (tbl.CurrentRow == lastRow)
					{
						txtResult.ForeColor = Color.Red;
						txtResult.AppendLine(str);
					}
				}
			};
			ZSyncMeli.Begin = (int aid, string name, string title) =>
			{
				ETRow eTRow = tbl.Rows.Add();
				eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE427F2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_833d9954b1984128a934e5888141bcde)] = aid.ToString();
				eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FB3C05 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5)] = name;
				eTRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E50E08 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce)] = title;
				tbl.Adjust();
			};
			ZSyncMeli.Process = (string str, string no) =>
			{
				ETRow lastRow = tbl.LastRow;
				if (lastRow != null)
				{
					lastRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61E50E08 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce)] = str;
					if (no.Length > 0)
					{
						string s = (string)(lastRow.Tag = ((lastRow.Tag is string text) ? (text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A929992 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380) + no) : no));
						if (tbl.CurrentRow == lastRow)
						{
							txtResult.AppendLine(s);
						}
					}
				}
			};
			ZSyncMeli.Percent = (int percent) =>
			{
				ETRow lastRow = tbl.LastRow;
				if (lastRow != null)
				{
					lastRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x643B10D2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848)] = percent + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0F11F7 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
				}
			};
			ZSyncMeli.Done = (int num, string msg) =>
			{
				ETRow lastRow = tbl.LastRow;
				if (lastRow != null)
				{
					ETCell eTCell = lastRow.Cells[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6A8ED58E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8a3ed341872c400cb534e2c2abe08c21)];
					eTCell.Color = ((num > 0) ? Color.DarkBlue : Color.DarkRed);
					eTCell.Text = num.ToString();
					lastRow[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x66512A8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f3d6deb4fde41d9a7761138db5819e8)] = msg;
					string s = (string)(lastRow.Tag = ((lastRow.Tag is string text) ? (text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x14BCED81 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_c108089651194500a647976b446524ef) + msg) : msg));
					if (tbl.CurrentRow == lastRow)
					{
						txtResult.AppendLine(s);
					}
				}
			};
			ZSyncMeli.Output = (string content) =>
			{
				txtResult.AppendText(content + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7086AAD8 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834));
			};
			ZSyncMeli.Recieve = (ZClient zc, string content) =>
			{
				txtResult.AppendText(content + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F214E92 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a));
			};
			ZSyncMeli.Finish = StartFinish;
			ZSyncMeli.Start(PStartDate.Value, PEndDate.Value, selectedItems);
		}
		catch (Exception ex)
		{
			StartFinish();
			EDialog.Show(ex);
		}
	}

	private void StartFinish()
	{
		try
		{
			btnStart.Enabled = true;
			btnStart.ForeColor = Color.Black;
			btnStart.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2ABB5C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3);
			lblDate.Text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -665719034 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_33cf50e8154f4189bf2233074109e835);
			任务.Instance.Disable(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x25A4875A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b5c61b0e89594879ac6181d2a6ea9010));
		}
		catch
		{
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
		this.lstAcc = new ZYing.UI.EListFilter();
		this.lblAll = new ZYing.UI.ELabel();
		this.body = new ZYing.UI.EControlCols();
		this.tbl = new ZYing.UI.ETable();
		this.txtResult = new ZYing.UI.ETextArea();
		this.top = new ZYing.UI.ELayer();
		this.btnStart = new ZYing.UI.EButton();
		this.PStartDate = new ZYing.UI.ESelectDate();
		this.lblDate = new ZYing.UI.ELabel();
		this.left = new ZYing.UI.ELayer();
		this.lblNote = new ZYing.UI.ELabel();
		this.main = new ZYing.UI.EContainer();
		this.PEndDate = new ZYing.UI.ESelectDate();
		this.body.SuspendLayout();
		this.top.SuspendLayout();
		this.left.SuspendLayout();
		this.main.SuspendLayout();
		base.SuspendLayout();
		this.lstAcc.ContainsKey = null;
		this.lstAcc.FlowAlign = ZYing.UI.FlowAlignment.Fill;
		this.lstAcc.IntegralHeight = false;
		this.lstAcc.Location = new System.Drawing.Point(3, 36);
		this.lstAcc.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.lstAcc.MaxHeight = 492;
		this.lstAcc.MinLines = 3;
		this.lstAcc.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x32243432 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		this.lstAcc.Padding = new System.Windows.Forms.Padding(4);
		this.lstAcc.RightMenu = null;
		this.lstAcc.SelectionMode = System.Windows.Forms.SelectionMode.MultiSimple;
		this.lstAcc.Size = new System.Drawing.Size(254, 492);
		this.lstAcc.TabIndex = 0;
		this.lblAll.FlowAlign = ZYing.UI.FlowAlignment.Right;
		this.lblAll.Location = new System.Drawing.Point(222, 3);
		this.lblAll.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x708785F4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_fc3cd536ccd84307b639bb91dda58834);
		this.lblAll.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.lblAll.Size = new System.Drawing.Size(35, 27);
		this.lblAll.TabIndex = 1;
		this.lblAll.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D342520 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
		this.body.Controls.Add(this.tbl);
		this.body.Controls.Add(this.txtResult);
		this.body.Dock = System.Windows.Forms.DockStyle.Fill;
		this.body.Location = new System.Drawing.Point(0, 36);
		this.body.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.body.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1DC8BEB2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_8085af1d9c424ce8858750a272deb5c5);
		this.body.Size = new System.Drawing.Size(634, 492);
		this.body.TabIndex = 3;
		this.body.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F22116A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_0f64bd204eb84a2ba2ced4f48511472a);
		this.tbl.IntegralHeight = false;
		this.tbl.LineStyle = System.Drawing.Drawing2D.DashStyle.Dot;
		this.tbl.Location = new System.Drawing.Point(0, 0);
		this.tbl.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.tbl.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6367C393 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3);
		this.tbl.ShowBorder = true;
		this.tbl.ShowHeader = false;
		this.tbl.Size = new System.Drawing.Size(312, 491);
		this.tbl.TabIndex = 2;
		this.txtResult.FlowAlign = ZYing.UI.FlowAlignment.Fill;
		this.txtResult.ForeColor = System.Drawing.Color.Red;
		this.txtResult.Location = new System.Drawing.Point(322, 0);
		this.txtResult.Margin = new System.Windows.Forms.Padding(3, 3, 3, 0);
		this.txtResult.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA0F5387 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd);
		this.txtResult.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.txtResult.Size = new System.Drawing.Size(312, 491);
		this.txtResult.TabIndex = 3;
		this.txtResult.TabStop = false;
		this.top.Controls.Add(this.btnStart);
		this.top.Controls.Add(this.PStartDate);
		this.top.Controls.Add(this.PEndDate);
		this.top.Controls.Add(this.lblDate);
		this.top.Dock = System.Windows.Forms.DockStyle.Top;
		this.top.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.top.Location = new System.Drawing.Point(0, 0);
		this.top.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x75CCD2AB ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1e8763c33f5c428eb6dda6fb6397a206);
		this.top.Padding = new System.Windows.Forms.Padding(0, 0, 0, 3);
		this.top.Size = new System.Drawing.Size(634, 36);
		this.top.TabIndex = 4;
		this.top.TextColor = System.Drawing.Color.Empty;
		this.btnStart.FlowBreak = false;
		this.btnStart.Location = new System.Drawing.Point(0, 3);
		this.btnStart.Margin = new System.Windows.Forms.Padding(0, 3, 3, 3);
		this.btnStart.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4078C2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f3e342896d674efba4d0e96d2c2c7835);
		this.btnStart.Padding = new System.Windows.Forms.Padding(8, 4, 8, 4);
		this.btnStart.Size = new System.Drawing.Size(130, 27);
		this.btnStart.TabIndex = 2;
		this.btnStart.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -65518943 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_3a5777d0c5a44cbeb8a0504656901848);
		this.PStartDate.AutoHeight = true;
		this.PStartDate.FlowBreak = false;
		this.PStartDate.Location = new System.Drawing.Point(136, 4);
		this.PStartDate.Margin = new System.Windows.Forms.Padding(3, 4, 3, 3);
		this.PStartDate.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2D32861E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_86f6039bdfae4fbf8b737ee958fa0d6b);
		this.PStartDate.NeedText = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A6B752A ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.PStartDate.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PStartDate.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x7C5E5BDA ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_896101e25d834efc9c6025c68e88b594);
		this.PStartDate.Size = new System.Drawing.Size(110, 26);
		this.PStartDate.TabIndex = 9;
		this.lblDate.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.lblDate.FlowBreak = false;
		this.lblDate.Location = new System.Drawing.Point(368, 3);
		this.lblDate.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -339600659 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_f7436d8a98f54ced838571b1146c8cda);
		this.lblDate.Size = new System.Drawing.Size(262, 27);
		this.lblDate.TabIndex = 3;
		this.left.Controls.Add(this.lblNote);
		this.left.Controls.Add(this.lblAll);
		this.left.Controls.Add(this.lstAcc);
		this.left.Dock = System.Windows.Forms.DockStyle.Left;
		this.left.Location = new System.Drawing.Point(12, 6);
		this.left.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x68D08ECE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_ba24c4ace9fb47e68caac58af31531d6);
		this.left.Padding = new System.Windows.Forms.Padding(0, 0, 5, 0);
		this.left.Size = new System.Drawing.Size(266, 529);
		this.left.TabIndex = 5;
		this.left.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x408BC5C6 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b310e74c706f46109e11eaa7d1ebbc44);
		this.left.TextColor = System.Drawing.Color.Empty;
		this.lblNote.FlowAlign = ZYing.UI.FlowAlignment.Justify;
		this.lblNote.FlowBreak = false;
		this.lblNote.Location = new System.Drawing.Point(3, 3);
		this.lblNote.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60FAB823 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5);
		this.lblNote.Padding = new System.Windows.Forms.Padding(0, 4, 6, 4);
		this.lblNote.Size = new System.Drawing.Size(213, 27);
		this.lblNote.TabIndex = 7;
		this.lblNote.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A973AC0 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_66e486d61d1a40bb93e94afae6d41380);
		this.main.Controls.Add(this.body);
		this.main.Controls.Add(this.top);
		this.main.Dock = System.Windows.Forms.DockStyle.Fill;
		this.main.Location = new System.Drawing.Point(278, 6);
		this.main.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22B3589E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_5799054b66cc446ab8f268aedb628cd2);
		this.main.Size = new System.Drawing.Size(635, 529);
		this.main.TabIndex = 6;
		this.main.Text = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4BB5009D ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_761e0fe709d244abb23c4e08203e7018);
		this.PEndDate.AutoHeight = true;
		this.PEndDate.FlowBreak = false;
		this.PEndDate.Location = new System.Drawing.Point(252, 4);
		this.PEndDate.Margin = new System.Windows.Forms.Padding(3, 4, 3, 3);
		this.PEndDate.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -789350866 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5);
		this.PEndDate.Padding = new System.Windows.Forms.Padding(6, 4, 6, 4);
		this.PEndDate.PlaceHolder = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x10C0DEA4 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_4798bae9c4c0490f88e66246f2f48f03);
		this.PEndDate.Size = new System.Drawing.Size(110, 26);
		this.PEndDate.TabIndex = 10;
		base.Controls.Add(this.main);
		base.Controls.Add(this.left);
		base.Name = NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A699094 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535);
		this.Padding = new System.Windows.Forms.Padding(12, 6, 12, 6);
		base.Size = new System.Drawing.Size(926, 542);
		this.body.ResumeLayout(false);
		this.top.ResumeLayout(false);
		this.top.PerformLayout();
		this.left.ResumeLayout(false);
		this.left.PerformLayout();
		this.main.ResumeLayout(false);
		base.ResumeLayout(false);
	}
}
