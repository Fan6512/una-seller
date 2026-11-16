using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Meli.Model;
using ZYing.Web;

namespace ZYing.Meli.Models.UserProducts;

public class CreateUserProductModel
{
	[CompilerGenerated]
	private sealed class <>c__DisplayClass61_0
	{
		public string id;

		public <>c__DisplayClass61_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <GetValueStruct>b__0(MeliAttribute x)
		{
			return x.id == id;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass65_0
	{
		public string id;

		public <>c__DisplayClass65_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <SetAttr>b__0(MeliAttribute x)
		{
			return x.id.EqualIgnore(id);
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass66_0
	{
		public string id;

		public <>c__DisplayClass66_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <SetAttrs>b__0(MeliAttribute x)
		{
			return x.id.EqualIgnore(id);
		}
	}

	public decimal? net_proceeds;

	public decimal? global_net_proceeds;

	public string family_name { get; set; }

	public string title { get; set; }

	public string category_id { get; set; }

	public decimal? price { get; set; }

	public int available_quantity { get; set; }

	public ProductDescription description { get; set; }

	[JsonIgnore]
	public string plain_text
	{
		get
		{
			return description?.plain_text;
		}
		set
		{
			if (!value.IsNullOrEmpty())
			{
				description = new ProductDescription
				{
					plain_text = value
				};
			}
		}
	}

	public string currency_id { get; set; }

	public string site_id { get; set; }

	public List<MeliAttribute> attributes { get; set; }

	public List<ItemPicuture> pictures { get; set; }

	public List<MeliAttribute> sale_terms { get; set; }

	public List<SiteToSell> sites_to_sell { get; set; }

	public List<MeliAttribute> VariationAttrs => attributes?.Where((MeliAttribute x) => x.AllowVariations).ToList();

	public void SetSaleTerm(string type, string time)
	{
		if (!type.IsNullOrEmpty())
		{
			sale_terms = new List<MeliAttribute>
			{
				new MeliAttribute
				{
					id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E55CC73 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4),
					value_id = type
				}
			};
			if (type != kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189F2422 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba))
			{
				sale_terms.Add(new MeliAttribute
				{
					id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31323F57 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1),
					value_name = time + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1C260EFE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21)
				});
			}
		}
	}

	public void SetCondition()
	{
		SetAttr(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6F9E2013 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_da7aea68c55141aaa23181fded21b782), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x1D3AD8D4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1851BB7F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233));
	}

	public string GetSellerSku()
	{
		return attributes?.FirstOrDefault((MeliAttribute x) => x.id.EqualIgnore(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7C67C490 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_db457e1da79640b7816cb6ccbfc1d724)))?.value_name;
	}

	public long GetGtinCode()
	{
		return (attributes?.FirstOrDefault((MeliAttribute x) => x.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x13123A4E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82a9ebeebdaa41eeb6924eb98e4d50c1))?.value_name?.ToLong()).GetValueOrDefault();
	}

	public int? GetWeight()
	{
		return (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C1B43B1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8853ae20538f41b5a858417f4442afe8)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F97BD36 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a)))?.GetWeight();
	}

	public int[] GetSize()
	{
		int num = (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDBAC1AC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x44367B72 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13)))?.GetCM() ?? 0;
		int num2 = (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FCEB57B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B2314D8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950)))?.GetCM() ?? 0;
		int num3 = (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2BA0446E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e92d645a8d04335b6bb7b1de9e0a791)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x399F0989 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21)))?.GetCM() ?? 0;
		if (num > 0 && num2 > 0 && num3 > 0)
		{
			return new int[3] { num, num2, num3 };
		}
		return null;
	}

	public string GetSizeGridId()
	{
		return attributes?.FirstOrDefault((MeliAttribute x) => x.id.EqualIgnore(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31FFEF6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339)))?.value_name;
	}

	public string GetSizeChartRowId()
	{
		return attributes?.FirstOrDefault((MeliAttribute x) => x.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E504E09 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4))?.value_name;
	}

	private MeliAttrValueStruct GetValueStruct(string id)
	{
		<>c__DisplayClass61_0 CS$<>8__locals2 = new <>c__DisplayClass61_0();
		CS$<>8__locals2.id = id;
		return (attributes?.FirstOrDefault((MeliAttribute x) => x.id == CS$<>8__locals2.id))?.ValueStruct;
	}

	public void SetGtinCode(string code)
	{
		SetAttr(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3139FF11 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1), code);
		attributes.RemoveAll((MeliAttribute x) => x.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F923B34 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a));
	}

	public void SetSellerSku(string sku)
	{
		SetAttr(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FBD978A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff), sku);
	}

	public void SetWeightSize(int weight, int[] size)
	{
		string id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA80D2A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22);
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(2, 1);
		defaultInterpolatedStringHandler.AppendFormatted(weight);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x399F09F1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21));
		SetAttr(id, defaultInterpolatedStringHandler.ToStringAndClear());
		string id2 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7347664 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3);
		defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(3, 1);
		defaultInterpolatedStringHandler.AppendFormatted(size[0]);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CDFFB88 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479));
		SetAttr(id2, defaultInterpolatedStringHandler.ToStringAndClear());
		string id3 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x110390FA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315);
		defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(3, 1);
		defaultInterpolatedStringHandler.AppendFormatted(size[1]);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FCEB5C7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913));
		SetAttr(id3, defaultInterpolatedStringHandler.ToStringAndClear());
		string id4 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C026283 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882);
		defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(3, 1);
		defaultInterpolatedStringHandler.AppendFormatted(size[2]);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-2066591320) ^ 0x2468E6FE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb));
		SetAttr(id4, defaultInterpolatedStringHandler.ToStringAndClear());
	}

	public void SetAttr(string id, string valueName, string valueId = null)
	{
		<>c__DisplayClass65_0 CS$<>8__locals3 = new <>c__DisplayClass65_0();
		CS$<>8__locals3.id = id;
		if (!valueName.IsNullOrEmpty() || !valueId.IsNullOrEmpty())
		{
			MeliAttribute meliAttribute = attributes?.FirstOrDefault((MeliAttribute x) => x.id.EqualIgnore(CS$<>8__locals3.id));
			if (meliAttribute == null)
			{
				meliAttribute = new MeliAttribute
				{
					id = CS$<>8__locals3.id
				};
				attributes.Add(meliAttribute);
			}
			meliAttribute.value_name = valueName;
			meliAttribute.value_id = valueId;
		}
	}

	public void SetAttrs(List<MeliAttribute> attrs)
	{
		foreach (MeliAttribute attr in attrs)
		{
			<>c__DisplayClass66_0 CS$<>8__locals3 = new <>c__DisplayClass66_0();
			if (!attr.value_name.IsNullOrEmpty() || !attr.value_id.IsNullOrEmpty())
			{
				CS$<>8__locals3.id = attr.id;
				MeliAttribute meliAttribute = attributes?.FirstOrDefault((MeliAttribute x) => x.id.EqualIgnore(CS$<>8__locals3.id));
				if (meliAttribute == null)
				{
					meliAttribute = new MeliAttribute
					{
						id = CS$<>8__locals3.id
					};
					attributes.Add(meliAttribute);
				}
				meliAttribute.value_name = attr.value_name;
				meliAttribute.value_id = attr.value_id;
			}
		}
	}

	public List<NAttr> BuildAttrs()
	{
		List<NAttr> list = new List<NAttr>();
		if (attributes == null || attributes.Count == 0)
		{
			return list;
		}
		foreach (MeliAttribute attribute in attributes)
		{
			if (attribute.id.In(MeliConsts.AttrNames.CommonAttrs))
			{
				continue;
			}
			NAttr nAttr = new NAttr
			{
				NameId = attribute.id,
				Name = attribute.name,
				ValueId = attribute.value_id.ToInt()
			};
			if (!attribute.value_name.IsNullOrEmpty())
			{
				if (!attribute.value_name.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FB63C02 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff)))
				{
					nAttr.Value = attribute.value_name;
				}
				else
				{
					string[] array = attribute.value_name.Split(' ');
					nAttr.Value = array[0];
					nAttr.Unit = array[1];
				}
			}
			list.Add(nAttr);
		}
		return list;
	}

	public string GetVariantKey()
	{
		List<string> list = VariationAttrs?.Select((MeliAttribute x) => x.value_name).ToList();
		if (list == null || list.Count == 0)
		{
			return "";
		}
		return string.Join(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -296711422 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4), list);
	}

	public CreateUserProductModel()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		currency_id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236683C0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8);
		site_id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE1B8147 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152);
		attributes = new List<MeliAttribute>();
		pictures = new List<ItemPicuture>();
		sale_terms = new List<MeliAttribute>();
		sites_to_sell = new List<SiteToSell>();
		base..ctor();
	}
}
