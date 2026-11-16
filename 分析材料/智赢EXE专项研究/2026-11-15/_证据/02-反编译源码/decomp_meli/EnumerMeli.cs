using System;
using System.Collections.Generic;
using System.Drawing;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing;
using ZYing.Data;
using ZYing.UI;

namespace Enumer;

public class Meli
{
	public enum NS
	{
		产品 = 1,
		图片,
		价格,
		库存
	}

	public enum Stat
	{
		待更新 = 1,
		待删除 = 2,
		失败 = 6,
		完成 = 7,
		零库存 = 8,
		已删除 = 9
	}

	public enum StatV2
	{
		活跃中 = 1,
		已暂停 = 2,
		已删除 = 3,
		失败 = 4,
		刊登中 = 5,
		被关闭 = 6,
		待审查 = 7,
		未激活 = 8,
		忽略 = 19
	}

	public enum QAStatus
	{
		已回答 = 1,
		未回答,
		已关闭,
		审查中,
		被禁用
	}

	public enum MeliTaskType
	{
		刊登产品 = 1,
		更新产品 = 2,
		删除产品 = 3,
		更新图片 = 4,
		更新价格 = 5,
		更新库存 = 6,
		更新重量尺寸 = 7,
		上架产品 = 8,
		下架产品 = 9,
		更新属性 = 10,
		更新描述 = 12,
		更新变体 = 13,
		拉取产品 = 14
	}

	public enum MeliTaskStatus
	{
		待处理 = 1,
		处理中 = 2,
		CBT完成 = 3,
		失败可重试 = 8,
		失败 = 9,
		完毕 = 10
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass30_0
	{
		public string key;

		public <>c__DisplayClass30_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <GetArea>b__0(SiteInfo i)
		{
			return i.Id == key;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass31_0
	{
		public string siteid;

		public <>c__DisplayClass31_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <GetCNName>b__0(SiteInfo i)
		{
			return i.Id == siteid;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass32_0
	{
		public string key;

		public <>c__DisplayClass32_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <GetCurByKey>b__0(SiteInfo i)
		{
			return i.Id == key;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass33_0
	{
		public string area;

		public <>c__DisplayClass33_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <GetSiteByArea>b__0(SiteInfo i)
		{
			return i.Area.EqualIgnore(area);
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass34_0
	{
		public string siteid;

		public <>c__DisplayClass34_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <FindSite>b__0(SiteInfo i)
		{
			return i.Id == siteid;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass35_0
	{
		public string siteid;

		public <>c__DisplayClass35_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <FindSiteCn>b__0(SiteInfo i)
		{
			return i.Id == siteid;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass36_0
	{
		public string key;

		public <>c__DisplayClass36_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <FindAbbrFullCnByKey>b__0(EItem i)
		{
			return i.Key == key;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass45_0
	{
		public string name;

		public <>c__DisplayClass45_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <GetStatV2>b__0(EItem x)
		{
			return x.Text == name;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass48_0
	{
		public string siteid;

		public <>c__DisplayClass48_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <GetItemLink>b__0(SiteInfo i)
		{
			return i.Id == siteid;
		}
	}

	private static readonly int[][] _stats;

	private static readonly List<SiteInfo> SiteList;

	public static readonly EItemCollection Items;

	public static readonly EItemCollection GroupAccs;

	public static readonly EItemCollection UPAccs;

	public static readonly EItemCollection AllCBTGroupAccs;

	public static readonly EItemCollection CBTAcc;

	public static readonly EItemCollection LocalAcc;

	public static readonly EItemCollection AllAcc;

	public static readonly EItemCollection Stats;

	public static readonly EItemCollection StatsV2;

	public static readonly EItemCollection Areas;

	public static readonly EItemCollection Abbrs;

	public static readonly EItemCollection Warrantys;

	public static readonly EItemCollection Sizes;

	public static readonly EItemCollection DefaultSizes;

	public static readonly EItemCollection QaStatus;

	public static readonly EItemCollection AccAbbrs;

	public static readonly EItemCollection LogTypes;

	static Meli()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		_stats = new int[4][]
		{
			new int[5] { 1, 2, 6, 7, 9 },
			new int[4] { 1, 2, 6, 7 },
			new int[3] { 1, 6, 7 },
			new int[4] { 1, 6, 7, 8 }
		};
		SiteList = new List<SiteInfo>
		{
			new SiteInfo
			{
				Id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0CBB5F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c),
				Area = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1AA931 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152),
				Currency = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x73F9B76 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3),
				Lan = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1272C633 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff),
				Zone = 3,
				Cn = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0DE41B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c),
				Name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x73EC434 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3),
				Url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FC4076B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913)
			},
			new SiteInfo
			{
				Id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x12739993 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff),
				Area = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1C2DE5AE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21),
				Currency = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA3E188 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22),
				Lan = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2B6E42 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed),
				Zone = 5,
				Cn = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8194EE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6),
				Name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76564920 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4),
				Url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F80CAA2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6)
			},
			new SiteInfo
			{
				Id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDB072BC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e),
				Area = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08EB5F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882),
				Currency = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F9D0D2E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a),
				Lan = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE78DB9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd),
				Zone = 4,
				Cn = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08EB57 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882),
				Name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2367B94E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8),
				Url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x310C82C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339)
			},
			new SiteInfo
			{
				Id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189F5AFC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba),
				Area = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x310F274 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339),
				Currency = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x5645F88F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120),
				Lan = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4ED20A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb),
				Zone = 5,
				Cn = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43E46453 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18),
				Name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1109231C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315),
				Url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4FFBE0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb)
			},
			new SiteInfo
			{
				Id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FC406FD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913),
				Area = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CD548A4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479),
				Currency = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443CC8DA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13),
				Lan = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46FBC911 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51),
				Zone = 3,
				Cn = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22024689 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2),
				Name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2367C668 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8),
				Url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C632A4A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)
			},
			new SiteInfo
			{
				Id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C118FDF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8853ae20538f41b5a858417f4442afe8),
				Area = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x217E3E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6ddd660130a6487b8fe1ae0feefd9bbe),
				Currency = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x185B7011 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233),
				Lan = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FB60D46 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff),
				Zone = 3,
				Cn = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1273E7A3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff),
				Name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1B88AF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152),
				Url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1D55BA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d)
			}
		};
		Items = new EItemCollection();
		GroupAccs = new EItemCollection();
		UPAccs = new EItemCollection();
		AllCBTGroupAccs = new EItemCollection();
		CBTAcc = new EItemCollection();
		LocalAcc = new EItemCollection();
		AllAcc = new EItemCollection();
		Stats = new EItemCollection();
		StatsV2 = new EItemCollection();
		Areas = new EItemCollection();
		Abbrs = new EItemCollection();
		Warrantys = new EItemCollection();
		Sizes = new EItemCollection();
		DefaultSizes = new EItemCollection();
		QaStatus = new EItemCollection();
		AccAbbrs = new EItemCollection();
		LogTypes = new EItemCollection();
		StatsV2.AddRange(new EItem(1, StatV2.活跃中.ToString(), Color.DarkGreen), new EItem(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1273E7DB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff), Color.Black), new EItem(4, StatV2.失败.ToString(), Color.DarkRed), new EItem(2, StatV2.已暂停.ToString(), Color.Blue), new EItem(3, StatV2.已删除.ToString(), Color.DarkOrchid), new EItem(7, StatV2.待审查.ToString(), Color.Red), new EItem(8, StatV2.未激活.ToString(), Color.Peru), new EItem(6, StatV2.被关闭.ToString(), Color.Peru), new EItem(19, StatV2.忽略.ToString(), Color.DarkGray));
		QaStatus.AddRange(new EItem(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1C2D7AE6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21), Color.DarkGreen), new EItem(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A3B79E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718), Color.Peru), new EItem(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1598F106 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120), Color.DarkGray), new EItem(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B29D9AA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950), Color.Red), new EItem(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F9D7000 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a), Color.DarkRed));
		Stats.AddRange(new EItem(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -161762668 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14), Color.Blue), new EItem(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x29DCEE30 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718), Color.DarkRed), new EItem(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2203E8F3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2)), new EItem(6, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443D1C82 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13), Color.Red), new EItem(7, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43E591D3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18)), new EItem(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46FB2B4B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51), Color.DarkMagenta), new EItem(9, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x352F0F16 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45), Color.DarkViolet));
		Warrantys.AddRange(new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08AF2F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E5F06C5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x131330FE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82a9ebeebdaa41eeb6924eb98e4d50c1))
		{
			Serial = 60
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDB00CFC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443CB650 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x417F8A4D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8))
		{
			Serial = 90
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7941B8E4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6F94EA69 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_da7aea68c55141aaa23181fded21b782), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A209A6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718)));
		Sizes.AddRange(new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x203492 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6ddd660130a6487b8fe1ae0feefd9bbe), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46FA9EB9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E5F0601 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE92E55 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2730BD99 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x705CC6C4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a21525dd11f740a2800d892a767a04e6)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CD47DB4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46FA9D29 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA2C340 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189E6F66 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x2EFDD149 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7941BBEE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE89E169 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x185B7255 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7C6D0ACA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_db457e1da79640b7816cb6ccbfc1d724)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x15995AA0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B93F693 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7941BA62 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D9C177 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA2C26A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76563486 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FB6F1F6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08AD71 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B29DB06 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -601960660 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8853ae20538f41b5a858417f4442afe8), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A0763EF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0bd1e5ac52394aa1bf047220642aab8c))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E5F0417 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x247E36FF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FC47AE5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0CC12B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x12724DCF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE860CF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x310B370 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C628090 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1C2CC1BC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x352EBF84 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45)
		});
		DefaultSizes.AddRange(new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x73EBE90 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x63BB8A55 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1B8CCD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B93F099 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x310B20E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C632F2A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4782C91D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3BB7408D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08AB81 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1B8D9B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1273E29B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A8D917 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08ABFF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08A80D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2730B9BB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189F23C0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D86D97 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443CB15E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3BB7437B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2730B935 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3995C359 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443CB1EE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08A8D1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x185B7799 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189F229C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FE54423 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C632D60 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x5B8603AC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2A3F58 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E5F00FF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1D5288 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F80B3DC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3170FACE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D86C29 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1B8039 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2A3FEA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE86CBF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x59AF0012 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x14973E83 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x310BF9C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60828929 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2BAA802E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e92d645a8d04335b6bb7b1de9e0a791)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3995CDC5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FB753C0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D86309 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d)
		}, new EItem(null, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B29D182 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4782C573 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D862AD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d)
		});
		Abbrs.AddRange(new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7940E214 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443CBE5E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3D109FA8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76563E70 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDB12C96 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2367B832 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A201F6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D86269 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D94B55 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x23322C72 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7D92706B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aa09e57b638a476287c30228f2ba6337))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08A7CB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE821DF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6082FE97 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76563E28 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B29D17C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6082C4D7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FB7244A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F80BD94 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46FA9695 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE81B7F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CD548B2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189F2C7C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -585884571 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3138F48D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7656378A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE6D1DD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0CCF2D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c)
		});
		AccAbbrs.AddRange(new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-813636989) ^ 0x6E213B80 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CD53D78 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -64396181 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1D5CB0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0DE599 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CD53D10 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0DE591 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F9D78D2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE69E99 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7B9D8224 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4d8504ccce2f465ba969392f5e065f21), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C11CAE9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8853ae20538f41b5a858417f4442afe8))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7941B09C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7941FC26 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x13133BDE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82a9ebeebdaa41eeb6924eb98e4d50c1), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A899A3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B29D152 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE81B7F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-813636989) ^ 0x493EAEB7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x233252CA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CD53D58 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4F8402 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08A4AB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189F250A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE6D1ED ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7D93594F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aa09e57b638a476287c30228f2ba6337)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43E42B87 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4F8EA4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA2C8EE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61265EE3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f8d88042fa6e4eae8f42d355fd0508fd)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1273EDD9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -1880382900 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1B8197 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61273AB9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f8d88042fa6e4eae8f42d355fd0508fd)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22024DA9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22024DCF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1273EEF3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1273A249 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A076B07 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0bd1e5ac52394aa1bf047220642aab8c), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x233259BE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4F8D8E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3D10E9DE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1B83B1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08A5A3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2730B59B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14)
		}, new EItem(World.GetIcon(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1B8843 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152)), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443CBC9C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D86025 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d))
		{
			Extra = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C63204C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)
		});
		LogTypes.AddRange(new EItem(1, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46FA9493 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51)), new EItem(2, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x867E32C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb)), new EItem(3, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7D93CFF3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aa09e57b638a476287c30228f2ba6337)), new EItem(4, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x21724C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6ddd660130a6487b8fe1ae0feefd9bbe)), new EItem(5, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7B9C3E6C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4d8504ccce2f465ba969392f5e065f21)), new EItem(6, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2B8E12 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed)), new EItem(7, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A8D0DD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd)), new EItem(8, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2A3410 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed)), new EItem(9, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -1399722727 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa)), new EItem(10, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA2CD04 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22)), new EItem(11, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x612773D3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f8d88042fa6e4eae8f42d355fd0508fd)), new EItem(12, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2367CA22 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8)), new EItem(13, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443CBB70 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13)), new EItem(14, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1B8465 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152)));
		foreach (SiteInfo site in SiteList)
		{
			if (site.Id != kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -983921270 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18))
			{
				Areas.Add(new EItem(site.Id, site.Cn));
			}
			EItem it = new EItem(World.GetIcon(site.Area), site.Id, site.Cn, site.Name)
			{
				Extra = site.Area
			};
			Items.Add(it);
		}
		LoadAsync().GetAwaiter();
	}

	public static async Task LoadAsync()
	{
		await MPush.GetAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x73B155C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3));
	}

	public static void Load(Record re)
	{
		List<EItem> list = new List<EItem>();
		List<EItem> list2 = new List<EItem>();
		while (re.Read())
		{
			int id = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2B7D9C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed));
			string text = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-2066591320) ^ 0x64AC4036 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6));
			string text2 = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x311F2D2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339));
			string text3 = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x23331712 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1));
			bool flag = re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x110950F0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315));
			string value = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x352FA6A6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45));
			bool flag2 = re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0CC973 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c));
			string text4 = string.Concat(str2: FindSiteCn(text) + (flag2 ? "" : kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FE54F85 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980)) + (flag ? kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1D598E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d) : ""), str0: text2.IsNullOrEmpty() ? text3 : text2, str1: kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE7BA8F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd), str3: kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2B7DD4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed));
			int num = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -1620998191 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df));
			EItem eItem = new EItem(id, text4)
			{
				Icon = World.GetIcon(GetArea(text)),
				Color = EColors.GetColor(value),
				Extra = text,
				Group = num,
				Serial = (flag2 ? 1 : 2)
			};
			int num2 = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x705D5E98 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a21525dd11f740a2800d892a767a04e6));
			if (re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FB6C1B2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff)) == 9)
			{
				eItem.Hide = true;
			}
			else if (Member.只看自己店铺)
			{
				if (Member.LoginID != num2)
				{
					eItem.Hide = true;
				}
			}
			else if (Member.SectID > 1)
			{
				EItem eItem2 = Login.Items.FindKey(num2);
				if (eItem2 == null || eItem2.Hide)
				{
					eItem.Hide = true;
				}
				else
				{
					eItem.Hide = false;
				}
			}
			else if (Member.Buys.Length != 0)
			{
				eItem.Hide = !num2.In(Member.Buys);
			}
			if (flag2)
			{
				list.Add(eItem);
			}
			else
			{
				list2.Add(eItem);
			}
		}
		CBTAcc.Update(list.OrderByDescending((EItem i) => i.Group).ToList());
		LocalAcc.Update(list2.OrderByDescending((EItem i) => i.Group).ToList());
		list.AddRange(list2);
		AllAcc.Update(list.OrderByDescending((EItem i) => i.Group).ToList());
		Parse(re);
	}

	private static void Parse(Record re)
	{
		GroupAccs.Clear();
		UPAccs.Clear();
		AllCBTGroupAccs.Clear();
		int num = 0;
		int type = 0;
		EItem eItem = null;
		while (re.Read())
		{
			int num2 = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE7BE27 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd));
			int id = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6F95AE73 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_da7aea68c55141aaa23181fded21b782));
			string text = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4EDE62 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb));
			string text2 = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236683A2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8));
			string text3 = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x866AD08 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb));
			string text4 = (text2.IsNullOrEmpty() ? text3 : text2);
			bool flag = re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443CBB48 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13));
			bool flag2 = re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2A3448 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed));
			Color color = EColors.GetColor(re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x73FA7D0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3)));
			re.Gets(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46FA93C1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51));
			int num3 = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C090281 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882));
			if (num2 != num || (MeliShop.IsUPMode(num3) && !MeliShop.IsUPMode(type)) || (!MeliShop.IsUPMode(num3) && MeliShop.IsUPMode(type)))
			{
				eItem = new EItem(num2, text4, color)
				{
					Extra = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2A349E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed)),
					Group = num2,
					Serial = (flag2 ? 1 : 2)
				};
				if (num3.In(4, 5))
				{
					UPAccs.Add(eItem);
				}
				else
				{
					GroupAccs.Add(eItem);
				}
				num = num2;
				type = num3;
			}
			Image icon = World.GetIcon(GetArea(text));
			string text5 = FindSiteCn(text) + (flag2 ? "" : kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3BB74963 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa)) + (flag ? kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x352EB7B0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45) : "");
			EItem eItem2 = new EItem(id, text5, color)
			{
				Icon = icon,
				Extra = text,
				Group = num2,
				Serial = num3,
				Tag = flag,
				Hide = !flag2
			};
			int num4 = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A946B9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd));
			if (re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-2066591320) ^ 0x5086BC3E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e92d645a8d04335b6bb7b1de9e0a791)) == 9)
			{
				eItem2.Hide = true;
			}
			else if (Member.只看自己店铺)
			{
				eItem2.Hide = Member.LoginID != num4;
			}
			else if (Member.SectID > 1)
			{
				EItem eItem3 = Login.Items.FindKey(num4);
				if (eItem3 == null || eItem3.Hide)
				{
					eItem2.Hide = true;
				}
			}
			else if (Member.Buys.Length != 0)
			{
				eItem2.Hide = !num4.In(Member.Buys);
			}
			if (!eItem2.Hide)
			{
				eItem.AddSubItem(eItem2);
			}
			eItem.SubItems = new EItemCollection(eItem.SubItems.OrderBy((EItem x) => x.Text));
			eItem.Sub = ((eItem.SubCount > 0) ? eItem.SubCount.ToString() : "");
			eItem.Hide = eItem.SubCount == 0;
		}
		GroupAccs.Update(GroupAccs.Where((EItem x) => !x.Hide).ToList());
		UPAccs.Update(UPAccs.Where((EItem x) => !x.Hide).ToList());
		AllCBTGroupAccs.AddRange(GroupAccs);
		AllCBTGroupAccs.AddRange(UPAccs);
		AllCBTGroupAccs.Update();
	}

	public static void SetStatV2(EControl ec, int stat)
	{
		EItem eItem = StatsV2.FindKey(stat);
		if (eItem == null)
		{
			ec.Text = "";
			return;
		}
		ec.ForeColor = eItem.Color;
		ec.Text = eItem.Text;
	}

	public static string GetArea(string key)
	{
		<>c__DisplayClass30_0 CS$<>8__locals2 = new <>c__DisplayClass30_0();
		CS$<>8__locals2.key = key;
		return SiteList.FirstOrDefault((SiteInfo i) => i.Id == CS$<>8__locals2.key)?.Area;
	}

	public static string GetCNName(string siteid)
	{
		<>c__DisplayClass31_0 CS$<>8__locals3 = new <>c__DisplayClass31_0();
		CS$<>8__locals3.siteid = siteid;
		return SiteList.FirstOrDefault((SiteInfo i) => i.Id == CS$<>8__locals3.siteid)?.Cn ?? CS$<>8__locals3.siteid;
	}

	public static string GetCurByKey(string key)
	{
		<>c__DisplayClass32_0 CS$<>8__locals2 = new <>c__DisplayClass32_0();
		CS$<>8__locals2.key = key;
		return SiteList.FirstOrDefault((SiteInfo i) => i.Id == CS$<>8__locals2.key)?.Currency;
	}

	public static string GetSiteByArea(string area)
	{
		<>c__DisplayClass33_0 CS$<>8__locals2 = new <>c__DisplayClass33_0();
		CS$<>8__locals2.area = area;
		return SiteList.FirstOrDefault((SiteInfo i) => i.Area.EqualIgnore(CS$<>8__locals2.area))?.Id;
	}

	public static SiteInfo FindSite(string siteid)
	{
		<>c__DisplayClass34_0 CS$<>8__locals2 = new <>c__DisplayClass34_0();
		CS$<>8__locals2.siteid = siteid;
		return SiteList.FirstOrDefault((SiteInfo i) => i.Id == CS$<>8__locals2.siteid);
	}

	public static string FindSiteCn(string siteid)
	{
		<>c__DisplayClass35_0 CS$<>8__locals2 = new <>c__DisplayClass35_0();
		CS$<>8__locals2.siteid = siteid;
		return SiteList.FirstOrDefault((SiteInfo i) => i.Id == CS$<>8__locals2.siteid)?.Cn;
	}

	public static string FindAbbrFullCnByKey(string key)
	{
		<>c__DisplayClass36_0 CS$<>8__locals2 = new <>c__DisplayClass36_0();
		CS$<>8__locals2.key = key;
		return AccAbbrs.FirstOrDefault((EItem i) => i.Key == CS$<>8__locals2.key)?.Text;
	}

	public static int Mask(string a, string c, string d, string e)
	{
		return Mask(Valid.ToInt(a), Valid.ToInt(c), Valid.ToInt(d), Valid.ToInt(e));
	}

	public static int Mask(int a, int c, int d, int e)
	{
		return a * 1000 + c * 100 + d * 10 + e;
	}

	private static int pow(int ns)
	{
		return (int)Math.Pow(10.0, 4 - ns);
	}

	public static int[] Union(NS ns, params Stat[] stats)
	{
		int[] array = new int[stats.Length];
		for (int i = 0; i < stats.Length; i++)
		{
			array[i] = (int)stats[i];
		}
		int[] a = ((ns == NS.产品) ? array : _stats[0]);
		int[] c = ((ns == NS.图片) ? array : _stats[1]);
		int[] d = ((ns == NS.价格) ? array : _stats[2]);
		int[] e = ((ns == NS.库存) ? array : _stats[3]);
		return Union(a, c, d, e);
	}

	private static int[] Union(int[] a, int[] c, int[] d, int[] e)
	{
		int[] array = new int[a.Length * c.Length * d.Length * e.Length];
		int num = 0;
		for (int i = 0; i < a.Length; i++)
		{
			for (int j = 0; j < c.Length; j++)
			{
				for (int k = 0; k < d.Length; k++)
				{
					for (int l = 0; l < e.Length; l++)
					{
						array[num++] = a[i] * 1000 + c[j] * 100 + d[k] * 10 + e[l];
					}
				}
			}
		}
		return array;
	}

	public static EItem GetStat(int stat, int ns)
	{
		int num = pow(ns);
		int id = stat % (num * 10) / num;
		return Stats.FindKey(id);
	}

	public static int? GetStat(string status)
	{
		if (!(status == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1273EB95 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff)))
		{
			if (!(status == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x23325E04 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1)))
			{
				if (!(status == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3138F829 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1)))
				{
					if (!(status == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2BAA841A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e92d645a8d04335b6bb7b1de9e0a791)))
					{
						if (!(status == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22024B83 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2)))
						{
							if (status == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA2CC36 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22))
							{
								return 6;
							}
							return null;
						}
						return 8;
					}
					return 6;
				}
				return 7;
			}
			return 2;
		}
		return 1;
	}

	public static EItem GetStatV2(int stat)
	{
		return StatsV2.FindKey(stat);
	}

	public static EItem GetStatV2(string name)
	{
		<>c__DisplayClass45_0 CS$<>8__locals2 = new <>c__DisplayClass45_0();
		CS$<>8__locals2.name = name;
		return StatsV2.Find((EItem x) => x.Text == CS$<>8__locals2.name);
	}

	public static string GetAcc(int id)
	{
		EItem eItem = AllAcc.FindKey(id);
		if (eItem != null)
		{
			return eItem.Text;
		}
		return "";
	}

	public static string GetAbbr(string accId)
	{
		EItem eItem = CBTAcc.FindKey(accId);
		if (eItem != null)
		{
			return eItem.Extra;
		}
		return "";
	}

	public static string GetItemLink(string itemid)
	{
		<>c__DisplayClass48_0 CS$<>8__locals3 = new <>c__DisplayClass48_0();
		if (itemid.IsNullOrEmpty() || itemid.Length < 3)
		{
			return "";
		}
		CS$<>8__locals3.siteid = itemid.Substring(0, 3);
		string value = itemid.Substring(3);
		string value2 = SiteList.FirstOrDefault((SiteInfo i) => i.Id == CS$<>8__locals3.siteid)?.Url;
		if (value2.IsNullOrEmpty())
		{
			return "";
		}
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(2, 3);
		defaultInterpolatedStringHandler.AppendFormatted(value2);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3170B900 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df));
		defaultInterpolatedStringHandler.AppendFormatted(CS$<>8__locals3.siteid);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6083AE2B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4));
		defaultInterpolatedStringHandler.AppendFormatted(value);
		return defaultInterpolatedStringHandler.ToStringAndClear();
	}

	public static string GetLan(string siteid)
	{
		if (!(siteid == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x51AE9A30 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff)))
		{
			if (siteid == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x185B0EC7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233))
			{
				return kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E5E26ED ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4);
			}
			return kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x352FED52 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45);
		}
		return kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A06327B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0bd1e5ac52394aa1bf047220642aab8c);
	}

	public static string GetImageId(string meliImgUrl)
	{
		if (meliImgUrl.IsNullOrEmpty())
		{
			return null;
		}
		if (meliImgUrl.IndexOf(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-713251553) ^ 0x662196ED ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22), StringComparison.OrdinalIgnoreCase) == -1)
		{
			return null;
		}
		string pattern = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A2052C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718);
		Match match = Regex.Match(meliImgUrl, pattern);
		if (!match.Success)
		{
			return null;
		}
		return match.Groups[1].Value;
	}

	public Meli()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}
}
