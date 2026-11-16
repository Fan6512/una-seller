using System;
using System.Collections.Concurrent;
using System.Collections.Specialized;
using System.Linq;
using System.Threading.Tasks;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Data;
using ZYing.Web;

namespace ZYing.Meli;

public class MeliCateAttrsHelper
{
	private static ConcurrentDictionary<string, MeliCateDto> cates;

	public static async Task<MeliCateDto> GetMeliCate(string site, int cateid)
	{
		if (site.IsNullOrEmpty() || cateid <= 0)
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x73206F6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3));
		}
		string key = site + cateid;
		if (!cates.TryGetValue(key, out var value))
		{
			ZClient zc = ZClient.Create(ZClient.美客多);
			zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794D058A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6);
			NameValueCollection nv = new NameValueCollection
			{
				[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794D05BE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)] = site,
				[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3139B15B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1)] = cateid.ToString()
			};
			await zc.PostAsync(nv);
			zc.Throw();
			Record record = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x12726C77 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff)];
			Record record2 = zc[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B9EA83D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c)];
			value = new MeliCateDto
			{
				Site = site,
				Id = cateid,
				Name = record.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x233EEFCC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1)),
				NameZh = record.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x38FC942A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04)),
				FullName = record.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x273C02BB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14)),
				FullZh = record.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B2565EC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950)),
				IsLeaf = record.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86B55B6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb)),
				ReferCbtId = record.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8C0908 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6)),
				Domain = record.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x127F5ABF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff)),
				MaxTitleLength = record.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C6F97D8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)),
				AllowVariations = record.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x220EFB5B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2)),
				MaxVariations = record.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C0413F1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882)),
				MaxPictures = record.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x127F5901 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff)),
				MaxPicturesVar = record.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -79080036 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)),
				UpdateAt = record.GetDate(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x127F5955 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff))
			};
			while (record2.Read())
			{
				MeliAttrDto item = new MeliAttrDto
				{
					Id = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE4DA13 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04)),
					Name = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A26869C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed)),
					NameZh = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CAE7FEE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22)),
					Required = record2.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D4D513 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d)),
					AllowVariation = record2.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443009D6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13)),
					VariationAttr = record2.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x608E3E95 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4)),
					Hidden = record2.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -2044517780 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315)),
					ValueType = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x608E3ED1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4)),
					ValueMaxLength = record2.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x18939A08 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba)),
					Multi = record2.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x18939A36 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba)),
					DefaultUnit = record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -233566488 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882)),
					Units = record2.Gets(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FBBE4C2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff))?.ToList(),
					Values = Json.ReadList<MeliIdName>(record2.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE14C837 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152)))
				};
				value.Attributes.Add(item);
			}
			cates[key] = value;
		}
		return value;
	}

	public static void ResetCateId(string site, int cateid)
	{
		cates.TryRemove($"{site}{cateid}", out var _);
	}

	public MeliCateAttrsHelper()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}

	static MeliCateAttrsHelper()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		cates = new ConcurrentDictionary<string, MeliCateDto>(StringComparer.OrdinalIgnoreCase);
	}
}
