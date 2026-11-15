using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using Enumer;
using Hmg1hUwjXmJQ0hx3wd;
using Logic;
using Xmbn6iXVDuARF6WhpW;
using ZYing.Data;
using ZYing.UI;
using ZYing.Web;

namespace ZYing;

public static class Member
{
	public const int 没有登陆 = -9;

	public const int 密码错误 = -10;

	public const int 管理1000 = 1000;

	public const int 培训中心 = 1801;

	public const int 杭州物流 = 1909;

	public const int 东莞仓库 = 5781;

	public const int 东莞美客多 = 1273;

	public const int 深圳胡辉 = 1795;

	public const int 智赢杭州 = 1117;

	public const int 测试号 = 1969;

	public const int 炬升 = 1023;

	public const int 青青 = 6238;

	public const int 测试 = 1051;

	public const int 产品部 = 2734;

	public const int 吕海鹏 = 1745;

	public const int 吕海鹏仓库 = 2941;

	public const int 自营跃科达 = 1002;

	public const int 老吕 = 1010;

	public const int 老张 = 1072;

	public const int 汽配王国 = 6422;

	public const int 速达腾辉 = 6470;

	public const int 万生科技 = 6606;

	public const int 徐俐斌 = 2117;

	public const int 跃科达1043 = 1043;

	public const int 高峰莱克奥斯 = 1892;

	public const int 海卖 = 4451;

	public const int 明远和邦 = 7613;

	public const int 保定Tik = 2722;

	public const int 古苇小小 = 4463;

	public const int 古苇小小_采集 = 9522;

	public const int 古苇海外仓 = 4566;

	public const int 以堂 = 6072;

	public const int 古苇墨西哥 = 6791;

	public const int 古苇欧西可森 = 6464;

	public const int 古小小长沙迈未 = 9475;

	public const int 美客多昭阅 = 5757;

	public const int 美客多铭淼 = 7102;

	public const int 美客多朝华商贸 = 6191;

	public const int 美客多斌斌 = 1526;

	public const int 张辉美咖 = 8883;

	public const int 小胖沈天为 = 1282;

	public const int 雷治国 = 1586;

	public const int 丰弘泰 = 1032;

	public const int 国贸eBay = 3760;

	public const int 总部 = 1;

	public const int 系统管理员 = 1;

	public const int 黄万生 = 56724;

	public const int 南野科技 = 1474;

	public static int[] 亚马逊守店人;

	public static DateTime ServerTime;

	public const int 八爪鱼 = 7445;

	public static string Access;

	public static string Secret;

	public static Guid Token;

	public static int LoginID;

	public static int SectID;

	public static string SectName;

	public static int UserID;

	private static int[] Sites;

	private static int[] AllowSites;

	public static int PSize;

	public static int PID;

	public static string Name;

	public static string Corp;

	public static string People;

	public static string BAppKey;

	public static string BSecret;

	public static string BToken;

	public static string AliKey;

	public static string AliSecret;

	public static string QQ;

	public static string IPLast;

	public static bool Depot;

	public static int Role;

	public static string RoleName;

	public static int Rank;

	public static Guid Auth;

	public static int Type;

	public static int Level;

	public static DateTime Ptime;

	public static string 公告;

	public static MConfig Config;

	public static string Head;

	public static int[] Buys;

	public static int[] Tags;

	public static readonly MLimit Limit;

	public static int 最大部门;

	public static bool 允许公告;

	public static bool 自己采购;

	public static bool 公开分销;

	public static bool 退货服务;

	public static bool 转运服务;

	public static bool 店铺提成;

	public static int 转运百分比;

	public static decimal 转运最高利润;

	public static decimal 转运贴标费;

	public static decimal 转运验货费;

	public static decimal 转运精检费;

	public static bool 保留For;

	public static bool 代缴费用;

	public static bool 批量采集;

	public static bool 产品为审核状态;

	public static bool 只看自己产品;

	public static bool 只看自己店铺;

	public static bool 欠费;

	public static bool 严重欠费;

	public const int 择远海外仓 = 5753;

	public const int 择远产品 = 3212;

	public const int 择远美客多 = 9667;

	internal static Member LhwDdyu1lYGkAdNUvQeo;

	public static bool Is美客多斌斌
	{
		get
		{
			if (!SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(UserID, new int[1] { 1526 }, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe))
			{
				return SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(PID, new int[1] { 1526 }, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe);
			}
			return true;
		}
	}

	public static bool Is南野科技
	{
		get
		{
			if (!SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(UserID, new int[1] { 1474 }, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe))
			{
				return SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(PID, new int[1] { 1474 }, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe);
			}
			return true;
		}
	}

	public static bool Is管理员 => Role == 1;

	public static bool Is系统管理员
	{
		get
		{
			if (Is管理员)
			{
				return SectID == 1;
			}
			return false;
		}
	}

	public static bool Is智赢官方转运
	{
		get
		{
			int userID = UserID;
			int[] array = new int[3];
			jgALHnPwz0q20IWSX4hh.RyQPtm98V8J(array, __ldtoken(<PrivateImplementationDetails>.75D2D4E930C5858904EB60F3DC891416128BD3222FFBB3A1457D6AC83171CD31), jgALHnPwz0q20IWSX4hh.mGmPMOsY0ou);
			return SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(userID, array, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe);
		}
	}

	public static bool Is明远
	{
		get
		{
			if (!SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(UserID, new int[1] { 7613 }, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe))
			{
				return SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(PID, new int[1] { 7613 }, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe);
			}
			return true;
		}
	}

	public static bool Is八爪鱼
	{
		get
		{
			if (UserID != 7445)
			{
				return PID == 7445;
			}
			return true;
		}
	}

	public static bool Is物流余额
	{
		get
		{
			if (!Is八爪鱼)
			{
				return Is择远用户;
			}
			return true;
		}
	}

	public static bool Is测试
	{
		get
		{
			if (UserID != 1051)
			{
				return PID == 1051;
			}
			return true;
		}
	}

	public static bool Is古苇小小
	{
		get
		{
			if (UserID != 4463)
			{
				return PID == 4463;
			}
			return true;
		}
	}

	public static bool Is张辉美咖
	{
		get
		{
			if (UserID != 8883)
			{
				return PID == 8883;
			}
			return true;
		}
	}

	public static bool Is万生科技
	{
		get
		{
			if (UserID != 6606)
			{
				return PID == 6606;
			}
			return true;
		}
	}

	public static bool 可以修改默认规则 => SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(LoginID, new int[1] { 972 }, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe);

	public static bool Is超级采集器
	{
		get
		{
			int userID = UserID;
			int[] array = new int[4];
			jgALHnPwz0q20IWSX4hh.RyQPtm98V8J(array, __ldtoken(<PrivateImplementationDetails>.40267BE5B7608D76AA6936FC8630CA0ACE680765EABC9EB13A615CEF3A10AFBD), jgALHnPwz0q20IWSX4hh.mGmPMOsY0ou);
			return SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(userID, array, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe);
		}
	}

	public static bool Is跟卖守店人 => SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(UserID, 亚马逊守店人, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe);

	public static bool Is跟卖守店人管理员
	{
		get
		{
			if (Is跟卖守店人 && AP4wNAu1mr569epDPLwC())
			{
				return SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(UserID, new int[2] { 5848, 1000 }, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe);
			}
			return false;
		}
	}

	public static bool IsTest => UserID == 1051;

	public static bool IsVIP => Level >= 9;

	public static int RootID
	{
		get
		{
			if (!IsRoot(PID))
			{
				return UserID;
			}
			return PID;
		}
	}

	public static string Joint => Logic.Joint.Encode(UserID);

	public static int SitesCount
	{
		get
		{
			if (!f1XGhrPiar5iDyBP8QdP.RyQPtm98V8J(Sites, f1XGhrPiar5iDyBP8QdP.Q6mPiNEvLQk))
			{
				return Sites.Length;
			}
			return 0;
		}
	}

	public static int AllowSitesCount
	{
		get
		{
			if (!f1XGhrPiar5iDyBP8QdP.RyQPtm98V8J(AllowSites, f1XGhrPiar5iDyBP8QdP.Q6mPiNEvLQk))
			{
				return AllowSites.Length;
			}
			return 0;
		}
	}

	public static string LevelName => Level switch
	{
		1 => A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x198476CD ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_725165346e184a6a9e1ca06eeca2c3c1), 
		2 => (string)pId70Pu1QQQqOCtLSdxF(0x7679CE55 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_1f0c339f2f5545dc882f485889cd9ab3), 
		3 => A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x57EC8128 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_aa754f91254c49a09d2da764295906ce), 
		4 => A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x94AC1F0 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_af81cda7f8cf4ede9fc6d65b3b0fd3ea), 
		5 => A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6A20B6B3 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_28721135bed84ee2ab8eccd2f8df395b), 
		6 => A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1572AD73 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_a5b9a804ca22442aa69baeecf1996b53), 
		9 => A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6D3A2348 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_7cb36848e8df46f6a04c61a49eb22da2), 
		10 => A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x39CBA13C ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_7dfb3c0237ac44a3ac16e4d95197f9dc), 
		77 => A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6B85E014 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_264717b9aada4491a008eb41e1264844), 
		88 => A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7EB224B3 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5d2b7fa25fb44ff4b3f647983f95e6d9), 
		_ => "", 
	};

	[Obsolete]
	public static bool Is择远新系统 => false;

	[Obsolete]
	public static bool Is择远产品
	{
		get
		{
			if (UserID != 3212)
			{
				return PID == 3212;
			}
			return true;
		}
	}

	public static bool Is择远管理 => UserID == 5753;

	public static bool Is择远转运
	{
		get
		{
			if (Is择远用户)
			{
				return 转运服务;
			}
			return false;
		}
	}

	public static bool Is择远用户 => Enumerable.Contains(Tags, 1);

	public static bool 可以代理产品 => Enumerable.Contains(Tags, 2);

	public static bool IsRoot()
	{
		return IsRoot((PID > 0) ? PID : UserID);
	}

	public static bool IsRoot(int uid)
	{
		if (Type != 5)
		{
			int[] array = new int[5];
			jgALHnPwz0q20IWSX4hh.RyQPtm98V8J(array, __ldtoken(<PrivateImplementationDetails>.57D3CC9763AAF437A654E67B1CB3A9BCB816CC5DB5A5FB0D3DFA0AFF94599394), jgALHnPwz0q20IWSX4hh.mGmPMOsY0ou);
			return SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(uid, array, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe);
		}
		return false;
	}

	public static bool 是关联子公司()
	{
		return PID > 0;
	}

	private static Guid gen(int id)
	{
		string text = HfS26pPiPodnnY1tC7XO.RyQPtm98V8J(HfS26pPiPodnnY1tC7XO.RRAPiuCjuQ6);
		StringBuilder stringBuilder = new StringBuilder(twsXv6Pt0TqWJgLIYbQo.RyQPtm98V8J(ref id, twsXv6Pt0TqWJgLIYbQo.VJ5PtcFXjFw));
		int num = 0;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5d0bda4c73b840a097775d00c98cbcda == 0)
		{
			num = 0;
		}
		switch (num)
		{
		default:
		{
			VKZssJPszl4gI385lG9s.RyQPtm98V8J(stringBuilder, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4E94C903 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_79a9a894f7a34841b4a43a973e6549cf), VKZssJPszl4gI385lG9s.l4KPWOZvJpr);
			for (int num2 = i97HtxPjymdGiLMljiyR.RyQPtm98V8J(text, i97HtxPjymdGiLMljiyR.CrZPjhYeUf2) - 1; num2 >= 0; num2--)
			{
				char c = soe6HcPjIlOe5a0XabHT.RyQPtm98V8J(text, num2, soe6HcPjIlOe5a0XabHT.rLXPj2mrJA9);
				glokPjPiK5HepPdfPpgD.RyQPtm98V8J(stringBuilder, num2, c, glokPjPiK5HepPdfPpgD.oxYPiLmxeCp);
			}
			if (B64GetPbrh7Uq6EdODXt.RyQPtm98V8J(stringBuilder, B64GetPbrh7Uq6EdODXt.wRkPbwUmHHd) > 32)
			{
				CaGtMrPilsat2YlwUsdr.RyQPtm98V8J(stringBuilder, 32, B64GetPbrh7Uq6EdODXt.RyQPtm98V8J(stringBuilder, B64GetPbrh7Uq6EdODXt.wRkPbwUmHHd) - 32, CaGtMrPilsat2YlwUsdr.fitPiFgBrYl);
			}
			return new Guid(Dx1n9ePWIL9h3apQMSPa.RyQPtm98V8J(stringBuilder, Dx1n9ePWIL9h3apQMSPa.ykxPW2uUQor));
		}
		}
	}

	public static void ParseNotice(Record re)
	{
		StringBuilder stringBuilder = new StringBuilder();
		while (srNvoRPMIM4DH0MDco10.RyQPtm98V8J(re, srNvoRPMIM4DH0MDco10.DInPM2cXxep))
		{
			string text = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, (string)pId70Pu1QQQqOCtLSdxF(0x79FEBE92 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_304a713a41a54d23b34fd4cec90492ed), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
			if (i97HtxPjymdGiLMljiyR.RyQPtm98V8J(text, i97HtxPjymdGiLMljiyR.CrZPjhYeUf2) != 0)
			{
				if (B64GetPbrh7Uq6EdODXt.RyQPtm98V8J(stringBuilder, B64GetPbrh7Uq6EdODXt.wRkPbwUmHHd) > 0)
				{
					s6MmkGPjP04kKZSFidCs.RyQPtm98V8J(stringBuilder, '\n', s6MmkGPjP04kKZSFidCs.nP0PjuhoZG7);
				}
				VKZssJPszl4gI385lG9s.RyQPtm98V8J(stringBuilder, text, VKZssJPszl4gI385lG9s.l4KPWOZvJpr);
				int num = 0;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_a5b9a804ca22442aa69baeecf1996b53 != 0)
				{
					num = 0;
				}
				switch (num)
				{
				}
			}
		}
		公告 = Dx1n9ePWIL9h3apQMSPa.RyQPtm98V8J(stringBuilder, Dx1n9ePWIL9h3apQMSPa.ykxPW2uUQor);
	}

	public static bool ContainSite(Func<int, bool> predicate)
	{
		if (f1XGhrPiar5iDyBP8QdP.RyQPtm98V8J(Sites, f1XGhrPiar5iDyBP8QdP.Q6mPiNEvLQk))
		{
			return true;
		}
		return Sites.Contains(predicate);
	}

	public static bool ContainSite(Site site)
	{
		return ContainSite((int)site);
	}

	public static bool ContainSite(params int[] sites)
	{
		if (f1XGhrPiar5iDyBP8QdP.RyQPtm98V8J(Sites, f1XGhrPiar5iDyBP8QdP.Q6mPiNEvLQk))
		{
			return true;
		}
		foreach (int num in sites)
		{
			if (mmHMS0PnPO67L2Z834X7.RyQPtm98V8J(Sites, num, mmHMS0PnPO67L2Z834X7.PhdPnuCqtKS))
			{
				return true;
			}
		}
		return false;
	}

	public static bool AllowSite(Site site)
	{
		return AllowSite((int)site);
	}

	public static bool AllowSite(params int[] sites)
	{
		if (AllowSites == null || AllowSites.Length == 0)
		{
			return ContainSite(sites);
		}
		int i = 0;
		int num = 0;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5d2b7fa25fb44ff4b3f647983f95e6d9 == 0)
		{
			num = 0;
		}
		switch (num)
		{
		}
		for (; i < sites.Length; i++)
		{
			int num2 = sites[i];
			if (mmHMS0PnPO67L2Z834X7.RyQPtm98V8J(AllowSites, num2, mmHMS0PnPO67L2Z834X7.PhdPnuCqtKS))
			{
				return true;
			}
		}
		return false;
	}

	public static void Parse(Record re)
	{
		Access = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x329BEB21 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_eb2c289411da4cb6ba956a784b15c8f8), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
		Secret = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x58090C9F ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_31cd0ff583b94a72adfb43c5b0e671c0), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
		LoginID = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0xD0D3145 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_c006ecad44af4f91bbb59ddc7246133d), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
		Token = YKlRrWPi4WVOvdEU5SCT.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7EB03ABB ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5d2b7fa25fb44ff4b3f647983f95e6d9), YKlRrWPi4WVOvdEU5SCT.tn0PimGtrFw);
		Sites = p8LRqGPUovaGP6On9JWy.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6CEB5BF7 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_8967b7ec79db46f5b434b4f13273f353), p8LRqGPUovaGP6On9JWy.XElPU6ntE8T);
		int num = 3;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_023613f232e14273961a9409b493fa36 == 0)
		{
			num = 3;
		}
		while (true)
		{
			switch (num)
			{
			case 1:
			{
				保留For = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, (string)pId70Pu1QQQqOCtLSdxF(0x19EF02A6 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5ed1c3cc2ae44eb0a7d1140d73c9845d), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				C4jxynPiQuAnM6IeCUAQ.RyQPtm98V8J(rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x304F4E39 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_dc6c739328c146a6a0bb367c92fb99b8), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC), C4jxynPiQuAnM6IeCUAQ.OotPikbG2eR);
				int userID = UserID;
				int[] array = new int[4];
				jgALHnPwz0q20IWSX4hh.RyQPtm98V8J(array, __ldtoken(<PrivateImplementationDetails>.09CCD6A8BD048AB478BF933F08B28E8F4400A1567FE55C3C1CBFF414401DE3F2), jgALHnPwz0q20IWSX4hh.mGmPMOsY0ou);
				if (SJ94w5PxXiKvvR8LvqbP.RyQPtm98V8J(userID, array, SJ94w5PxXiKvvR8LvqbP.xbCPxgvcqEe))
				{
					EImage.允许显示 = true;
				}
				VKla3du1klMDGMOtPe26();
				return;
			}
			case 6:
				Level = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1573BBED ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_a5b9a804ca22442aa69baeecf1996b53), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				Auth = YKlRrWPi4WVOvdEU5SCT.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4E94CA3D ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_79a9a894f7a34841b4a43a973e6549cf), YKlRrWPi4WVOvdEU5SCT.tn0PimGtrFw);
				Ptime = hZbRYrPtSO7HIBAmvwCV.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1573A7AB ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_a5b9a804ca22442aa69baeecf1996b53), hZbRYrPtSO7HIBAmvwCV.TSePtJRI47c);
				Head = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x62AFC93A ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_fa27fd6b6f2a44f6a9954a4ca58fab8d), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				Buys = p8LRqGPUovaGP6On9JWy.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3ADA5C57 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_63793a129ada454bbfdde45cfbec4aa2), p8LRqGPUovaGP6On9JWy.XElPU6ntE8T);
				Tags = p8LRqGPUovaGP6On9JWy.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4E94CA57 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_79a9a894f7a34841b4a43a973e6549cf), p8LRqGPUovaGP6On9JWy.XElPU6ntE8T);
				Limit.Parse(rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, (string)pId70Pu1QQQqOCtLSdxF(0x1FD326E ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_582502ab2572468fb1525174a0f0f129), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC));
				最大部门 = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4C248CC ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_85d2b8790f40427d85672cf528d7a58e), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				Config = Json.Read<MConfig>(rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x79FEBDDC ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_304a713a41a54d23b34fd4cec90492ed), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC));
				允许公告 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-(-326985204) ^ 0x1A36AE50 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_af81cda7f8cf4ede9fc6d65b3b0fd3ea), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				QQ = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4E96C2E7 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_79a9a894f7a34841b4a43a973e6549cf), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				IPLast = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x19ECE36A ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5ed1c3cc2ae44eb0a7d1140d73c9845d), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				自己采购 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4E94CA87 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_79a9a894f7a34841b4a43a973e6549cf), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				公开分销 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, (string)pId70Pu1QQQqOCtLSdxF(0x6CEB5839 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_8967b7ec79db46f5b434b4f13273f353), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				退货服务 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3298FD11 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_eb2c289411da4cb6ba956a784b15c8f8), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				num = 0;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_a6b3ee98172d4292a6820ea13f0df8f2 == 0)
				{
					num = 0;
				}
				break;
			case 2:
				Depot = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x61D958C3 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f61e9b55cd0f4e129854e3f8154e36b2), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				Rank = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x72EA0B35 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_d84cda0b6f1b48c48af618b969d85ac0), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				num = 5;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_d07f2917bfe649d6bc2f7fe0a441c80e != 0)
				{
					num = 3;
				}
				break;
			case 5:
				Type = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4C1B9D6 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_85d2b8790f40427d85672cf528d7a58e), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				num = 6;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_d3889fa4e0964f4e92c76625d59bfec4 != 0)
				{
					num = 3;
				}
				break;
			case 4:
				BAppKey = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-1506365678 ^ -1298650824 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_a6b3ee98172d4292a6820ea13f0df8f2), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				BSecret = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x140B769D ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_b6b2510e4e2944ba98b246aac342307e), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				BToken = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x502ED60D ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_6047c4a60e7b4ce3bb5b0bd793ab5b8c), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				AliKey = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1573A799 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_a5b9a804ca22442aa69baeecf1996b53), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				AliSecret = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6CEB5881 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_8967b7ec79db46f5b434b4f13273f353), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				num = 2;
				break;
			case 3:
				AllowSites = p8LRqGPUovaGP6On9JWy.RyQPtm98V8J(re, (string)pId70Pu1QQQqOCtLSdxF(-1506365678 ^ -1436437163 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f6823ca2c2d44cdba0ce31084c91362a), p8LRqGPUovaGP6On9JWy.XElPU6ntE8T);
				PSize = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x64DDCEDC ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_dd6240aeb46e481f98298eeaa3803307), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				if (PSize == 0)
				{
					PSize = 1600;
				}
				SectID = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x5CD639EE ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_c3dfc00b7eb944178b6ec529d7cbfa10), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				SectName = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0xD0E2499 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_c006ecad44af4f91bbb59ddc7246133d), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				PID = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x438AE95E ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_adf85748c2d94a4ab62195edb35b4b8c), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				UserID = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x57EE9DAA ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_aa754f91254c49a09d2da764295906ce), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				Name = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-922330359 ^ -1253982441 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_36369014ecd545c39ede59c60a92f687), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				Corp = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x15704045 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_a5b9a804ca22442aa69baeecf1996b53), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				People = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x9F03958 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f9db7669d8794e939d21bb92177af0a3), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				Role = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4C0405A ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_85d2b8790f40427d85672cf528d7a58e), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				RoleName = rLcSkWPt8yCVIq7PmbZo.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2D165892 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_69abf8cc2cd9444db21ebabb4b202604), rLcSkWPt8yCVIq7PmbZo.FxDPtYgHkBC);
				if (i97HtxPjymdGiLMljiyR.RyQPtm98V8J(RoleName, i97HtxPjymdGiLMljiyR.CrZPjhYeUf2) == 0)
				{
					RoleName = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-1506365678 ^ -983870777 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_c7210e071b3e401da22af49080b21ab3);
					num = 1;
					if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5d2b7fa25fb44ff4b3f647983f95e6d9 != 0)
					{
						num = 4;
					}
					break;
				}
				goto case 4;
			default:
				转运服务 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x9F3DE54 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f9db7669d8794e939d21bb92177af0a3), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				店铺提成 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x14AE8534 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_a6b3ee98172d4292a6820ea13f0df8f2), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				转运百分比 = SJDldcPt7tcODn1NdCRm.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0xC5719D1 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f6823ca2c2d44cdba0ce31084c91362a), SJDldcPt7tcODn1NdCRm.I7rPtAn6i4G);
				转运最高利润 = byxNyJPUtCKQmv4pX9Zq.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2426A518 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_0a5bf5790f514d7aaedef3a46965d602), byxNyJPUtCKQmv4pX9Zq.YEpPUsLrZ8S);
				转运贴标费 = byxNyJPUtCKQmv4pX9Zq.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x329A9171 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_eb2c289411da4cb6ba956a784b15c8f8), byxNyJPUtCKQmv4pX9Zq.YEpPUsLrZ8S);
				转运验货费 = byxNyJPUtCKQmv4pX9Zq.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3ADA39E3 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_63793a129ada454bbfdde45cfbec4aa2), byxNyJPUtCKQmv4pX9Zq.YEpPUsLrZ8S);
				转运精检费 = byxNyJPUtCKQmv4pX9Zq.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x329A90AD ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_eb2c289411da4cb6ba956a784b15c8f8), byxNyJPUtCKQmv4pX9Zq.YEpPUsLrZ8S);
				代缴费用 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x254957D0 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_67e3127100984868bbd1532c750888cc), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				批量采集 = Level >= 4 || d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7B58AC2D ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_27dc7b16621540c3867a9b2973100be1), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				产品为审核状态 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x61DB50E3 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f61e9b55cd0f4e129854e3f8154e36b2), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				只看自己产品 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x949C26E ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_af81cda7f8cf4ede9fc6d65b3b0fd3ea), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				只看自己店铺 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2424ACBE ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_0a5bf5790f514d7aaedef3a46965d602), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				欠费 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, (string)pId70Pu1QQQqOCtLSdxF(0x254957E4 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_67e3127100984868bbd1532c750888cc), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				严重欠费 = d9TtB4PsV9SFX72GZVKl.RyQPtm98V8J(re, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x19ECE2A4 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5ed1c3cc2ae44eb0a7d1140d73c9845d), d9TtB4PsV9SFX72GZVKl.Q4fPsdxTydk);
				num = 1;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_0fdc314caa8747e9a108f22fc8bec3bd != 0)
				{
					num = 0;
				}
				break;
			}
		}
	}

	public static string GenWebToken(DateTime? expire)
	{
		DateTime? dateTime = expire;
		DateTime dateTime3;
		if (!dateTime.HasValue)
		{
			DateTime dateTime2 = A0X3HkPtQh8HhvIRv9Y9.RyQPtm98V8J(A0X3HkPtQh8HhvIRv9Y9.GhQPtkN7xja);
			dateTime3 = Xx4muaPWV2Ev5RsAxL3l.RyQPtm98V8J(ref dateTime2, 2.0, Xx4muaPWV2Ev5RsAxL3l.LllPWd7l7P7);
		}
		else
		{
			dateTime3 = dateTime.GetValueOrDefault();
		}
		DateTime dateTime4 = dateTime3;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(1, 2);
		ws9EKEPWQnZ3qCKkadKV.RyQPtm98V8J(ref defaultInterpolatedStringHandler, V5TgLnPiZloSPhpxqqdx.RyQPtm98V8J(Token, V5TgLnPiZloSPhpxqqdx.rFyPiBPGSRS), ws9EKEPWQnZ3qCKkadKV.VtvPWkB3BtN);
		bSdHi8PWZxkkPUQpQn0y.RyQPtm98V8J(ref defaultInterpolatedStringHandler, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x140B741B ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_b6b2510e4e2944ba98b246aac342307e), bSdHi8PWZxkkPUQpQn0y.XuBPWBTbL7n);
		defaultInterpolatedStringHandler.AppendFormatted(oSKjMRPsqZD0dnjUwBhk.RyQPtm98V8J(ref dateTime4, oSKjMRPsqZD0dnjUwBhk.GXFPsvOyqRC));
		return GswD4BPitxStXhXeSFn6.RyQPtm98V8J(rOh2jiPiqVjYkxAs72FA.RyQPtm98V8J(e97dHCPWqAhbJQ9NRY5l.RyQPtm98V8J(ref defaultInterpolatedStringHandler, e97dHCPWqAhbJQ9NRY5l.a0uPWvp3CJW), rOh2jiPiqVjYkxAs72FA.FMNPivgYymE), null, GswD4BPitxStXhXeSFn6.DyWPisMnpON);
	}

	private static void SetConfig()
	{
		if (Is智赢官方转运)
		{
			Config current = JsonConfig<ZYing.Config>.Current;
			current.HongKongDownLoad = true;
			current.Save();
		}
	}

	public static void Logout()
	{
		HashSet<ZClient.IPGroup> hashSet = new HashSet<ZClient.IPGroup>
		{
			ZClient.Server,
			ZClient.老吕,
			ZClient.刊登,
			ZClient.香港,
			ZClient.网站,
			ZClient.速卖通
		};
		string value = Token.ToString();
		foreach (ZClient.IPGroup item in hashSet)
		{
			ZClient zClient = ZClient.Create(item);
			zClient.Command = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6CEB59CF ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_8967b7ec79db46f5b434b4f13273f353);
			zClient.Failed = (Exception ex) =>
			{
			};
			zClient.PostAsync(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x636DFF6B ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_c7210e071b3e401da22af49080b21ab3), value).GetAwaiter();
		}
		Token = Guid.Empty;
	}

	static Member()
	{
		A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		A8RedEce1ZSTaHmUDS.vv9ejnCAD();
		R7c2WOBSNAiVDsbUZ4.u2ulMj8ByVd();
		int[] array = new int[7];
		jgALHnPwz0q20IWSX4hh.RyQPtm98V8J(array, __ldtoken(<PrivateImplementationDetails>.CF9890E11C9376BF1B0D26274094F853F5562313EF459125CDE427320E7EE6BF), jgALHnPwz0q20IWSX4hh.mGmPMOsY0ou);
		亚马逊守店人 = array;
		SectName = "";
		Name = "";
		Corp = "";
		People = "";
		BAppKey = "";
		BSecret = "";
		BToken = "";
		AliKey = "";
		AliSecret = "";
		int num = 0;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f61e9b55cd0f4e129854e3f8154e36b2 != 0)
		{
			num = 0;
		}
		while (true)
		{
			switch (num)
			{
			case 2:
				转运服务 = false;
				店铺提成 = false;
				num = 1;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_121f7d146ff746ea9eb7be772374bcd9 == 0)
				{
					num = 1;
				}
				continue;
			case 1:
				return;
			}
			QQ = "";
			IPLast = "";
			RoleName = "";
			公告 = "";
			Config = new MConfig();
			Head = "";
			Buys = new int[0];
			Tags = new int[0];
			Limit = new MLimit();
			最大部门 = 5;
			允许公告 = false;
			自己采购 = false;
			公开分销 = false;
			退货服务 = false;
			num = 2;
			if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_3663c4b1d29e41d1b96ddfa6e70681bd != 0)
			{
				num = 1;
			}
		}
	}

	internal static bool CNq2Squ1F5ZZaNLmk1U0()
	{
		return LhwDdyu1lYGkAdNUvQeo == null;
	}

	internal static Member SKfFlVu14o0jgJr9WOfD()
	{
		return LhwDdyu1lYGkAdNUvQeo;
	}

	internal static bool AP4wNAu1mr569epDPLwC()
	{
		return Is系统管理员;
	}

	internal static object pId70Pu1QQQqOCtLSdxF(int P_0)
	{
		return A8RedEce1ZSTaHmUDS.DQmUrYZ9D(P_0);
	}

	internal static void VKla3du1klMDGMOtPe26()
	{
		SetConfig();
	}
}
