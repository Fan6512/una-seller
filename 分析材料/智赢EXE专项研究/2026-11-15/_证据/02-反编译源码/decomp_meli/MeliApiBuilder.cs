using System.Collections.Concurrent;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Threading.Tasks;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;

namespace ZYing.Meli;

public class MeliApiBuilder
{
	private static ConcurrentDictionary<int, MeliApi> _apis;

	public static async Task<MeliApi> Create(int aid = 0)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(1, 2);
		defaultInterpolatedStringHandler.AppendFormatted(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EEA6789 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd));
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31719BB0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df));
		defaultInterpolatedStringHandler.AppendFormatted(aid);
		using (await AsyncLocker.LockAsync(defaultInterpolatedStringHandler.ToStringAndClear()))
		{
			if (aid <= 0)
			{
				return await GetFirst();
			}
			if (_apis.TryGetValue(aid, out var value))
			{
				return value;
			}
			ZClient zc = ZClient.Create(ZClient.美客多);
			zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x15944C64 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120);
			await zc.PostAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FC53CB5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913), aid);
			zc.Throw();
			value = new MeliApi(zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43E5AA65 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18)]);
			_apis[aid] = value;
			return value;
		}
	}

	private static async Task<MeliApi> GetFirst()
	{
		MeliApi meliApi = _apis.Values.FirstOrDefault();
		if (meliApi != null && !meliApi.TokenExpired())
		{
			return meliApi;
		}
		ZClient zc = ZClient.Create(ZClient.美客多);
		zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A46049 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd);
		await zc.PostAsync();
		zc.Throw();
		MeliApi meliApi2 = new MeliApi(zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA34A74 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22)]);
		_apis[meliApi2.Aid] = meliApi2;
		return meliApi2;
	}

	public MeliApiBuilder()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}

	static MeliApiBuilder()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		_apis = new ConcurrentDictionary<int, MeliApi>();
	}
}
