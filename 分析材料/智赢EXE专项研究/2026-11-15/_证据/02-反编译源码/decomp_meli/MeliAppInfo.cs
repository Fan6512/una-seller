using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Threading.Tasks;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Data;

namespace ZYing.Meli;

public class MeliAppInfo
{
	[CompilerGenerated]
	private sealed class <>c__DisplayClass41_0
	{
		public string appId;

		public <>c__DisplayClass41_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <GetByAppId>b__0(MeliAppInfo x)
		{
			return x.AppId == appId;
		}
	}

	private static List<MeliAppInfo> AppList;

	public string AppName { get; set; }

	public string AppId { get; set; }

	public string AppSecret { get; set; }

	public string RedirectUrl { get; set; }

	public bool IsCbt { get; set; }

	public bool IsSelf { get; set; }

	public static MeliAppInfo ZYing => new MeliAppInfo
	{
		AppName = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x185CB8AB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233),
		AppId = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x273A9E49 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14),
		AppSecret = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7B97AA7A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4d8504ccce2f465ba969392f5e065f21),
		RedirectUrl = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794B9B00 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)
	};

	public static MeliAppInfo PuHui => new MeliAppInfo
	{
		AppName = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4788EEDF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c),
		AppId = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x73498F0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3),
		AppSecret = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3132D63D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1),
		RedirectUrl = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46F0BC6D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51)
	};

	public static MeliAppInfo JuYing => new MeliAppInfo
	{
		AppName = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7056E7B0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a21525dd11f740a2800d892a767a04e6),
		AppId = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x233871B8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1),
		AppSecret = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A2FF5F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd),
		RedirectUrl = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1705051699 ^ -906230262 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd)
	};

	public static MeliAppInfo Cohere => new MeliAppInfo
	{
		AppName = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A0D438B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0bd1e5ac52394aa1bf047220642aab8c),
		AppId = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FBD78DE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff),
		AppSecret = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FEF6701 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980),
		RedirectUrl = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7349958 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3)
	};

	[Obsolete("之前的本土号后台登录不进去了，切换到 本土ZYing")]
	private static MeliAppInfo 本土 => new MeliAppInfo
	{
		AppName = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x23359AF2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1),
		AppId = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B99D1C1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c),
		AppSecret = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7D985DAB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aa09e57b638a476287c30228f2ba6337),
		RedirectUrl = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8A9630 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6),
		IsCbt = false
	};

	private static MeliAppInfo 本土ZYing => new MeliAppInfo
	{
		AppName = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7C672C94 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_db457e1da79640b7816cb6ccbfc1d724),
		AppId = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1279C3E5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff),
		AppSecret = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1771EC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d),
		RedirectUrl = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE11AB37 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152),
		IsCbt = false
	};

	static MeliAppInfo()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		AppList = new List<MeliAppInfo> { ZYing, PuHui, JuYing, Cohere, 本土ZYing, 本土 };
		Load().GetAwaiter();
	}

	public static async Task Load()
	{
		ZClient zClient = ZClient.Create(ZClient.美客多);
		zClient.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A0BDC47 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0bd1e5ac52394aa1bf047220642aab8c);
		zClient.Finish = (ZClient z) =>
		{
			Record record = z[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189EAEDE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba)];
			if (record.Count > 0)
			{
				SetSelfApp(record);
			}
		};
		await zClient.PostAsync();
	}

	public static MeliAppInfo GetAppInfo(bool isCbt)
	{
		if (isCbt)
		{
			MeliAppInfo meliAppInfo = GetSelfApp();
			if (meliAppInfo == null)
			{
				meliAppInfo = (Member.Is古苇小小 ? Cohere : ZYing);
			}
			return meliAppInfo;
		}
		return 本土ZYing;
	}

	public static MeliAppInfo GetAppInfo(string siteId)
	{
		return GetAppInfo(siteId.Equals(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x76563E98 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4), StringComparison.OrdinalIgnoreCase));
	}

	public static MeliAppInfo GetByAppId(string appId)
	{
		<>c__DisplayClass41_0 CS$<>8__locals2 = new <>c__DisplayClass41_0();
		CS$<>8__locals2.appId = appId;
		return AppList.FirstOrDefault((MeliAppInfo x) => x.AppId == CS$<>8__locals2.appId);
	}

	public static MeliAppInfo GetSelfApp()
	{
		return AppList.FirstOrDefault((MeliAppInfo x) => x.IsSelf);
	}

	public static void SetSelfApp(Record re)
	{
		MeliAppInfo item = new MeliAppInfo
		{
			AppId = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31719BBA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df)),
			AppSecret = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8A9142 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6)),
			AppName = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE241C5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04)),
			RedirectUrl = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EECFE95 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd)),
			IsSelf = true
		};
		RemoveSelf();
		AppList.Add(item);
	}

	public static void RemoveSelf()
	{
		AppList.RemoveAll((MeliAppInfo x) => x.IsSelf);
	}

	public MeliAppInfo()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		IsCbt = true;
		base..ctor();
	}
}
