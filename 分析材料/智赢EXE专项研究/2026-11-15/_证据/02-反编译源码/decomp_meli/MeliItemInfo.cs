using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;

namespace ZYing.Meli.Model;

public class MeliItemInfo : MeliErrorCause
{
	public class SellerAddress
	{
		public class MCity
		{
			public string id;

			public string name;

			public MCity()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				id = "";
				name = "";
				base..ctor();
			}
		}

		public class MState
		{
			public string id;

			public string name;

			public MState()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				id = "";
				name = "";
				base..ctor();
			}
		}

		public class MCountry
		{
			public string id;

			public string name;

			public MCountry()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				id = "";
				name = "";
				base..ctor();
			}
		}

		public class MSearch_location
		{
			public MCity city;

			public MState state;

			public MSearch_location()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public MCity city;

		public MState state;

		public MCountry country;

		public MSearch_location search_location;

		public long id;

		public SellerAddress()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	public class ItemInfoDesc
	{
		public string id;

		public ItemInfoDesc()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	public class ItemRelation
	{
		public string id;

		public long? variation_id;

		public int? stock_relation;

		public ItemRelation()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	public class NetProceeds
	{
		public decimal amount { get; set; }

		public string currency_id { get; set; }

		public List<AdditionalConcept> additional_concepts { get; set; }

		public decimal ShipCost => (additional_concepts?.FirstOrDefault((AdditionalConcept i) => i.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8F0DD2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6))?.amount).GetValueOrDefault();

		public decimal SaleFee => (additional_concepts?.FirstOrDefault((AdditionalConcept i) => i.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x18909C42 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba))?.amount).GetValueOrDefault();

		public NetProceeds()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	public class AdditionalConcept
	{
		public const string ShippingCost = "shipping_cost";

		public const string SaleFee = "sale_fee";

		public string id { get; set; }

		public decimal amount { get; set; }

		public AdditionalConcept()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass74_0
	{
		public long? variationid;

		public string sellerSku;

		public <>c__DisplayClass74_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <FindImgUrl>b__0(Variation i)
		{
			return i.id == variationid;
		}

		internal bool <FindImgUrl>b__1(Variation i)
		{
			return i.GetSellerSku() == sellerSku;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass74_1
	{
		public string picId;

		public <>c__DisplayClass74_1()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <FindImgUrl>b__2(ItemPicuture i)
		{
			return i.id == picId;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass79_0
	{
		public string id;

		public <>c__DisplayClass79_0()
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

	public string id;

	public string item_id;

	public string cbt_item_id;

	public string user_product_id;

	public string seller_id;

	public string condition;

	public long? official_store_id;

	public string currency_id;

	public string site_id;

	public string category_id;

	public string domain_id;

	public string inventory_id;

	public int initial_quantity;

	public int? sold_quantity;

	public int? available_quantity;

	public List<ItemInfoDesc> descriptions;

	public SellerAddress seller_address;

	public string listing_type_id;

	public DateTime? last_updated;

	public DateTime? date_created;

	public DateTime? start_time;

	public DateTime? stop_time;

	public DateTime? end_time;

	public string status;

	public List<string> sub_status;

	public string permalink;

	public string thumbnail_id;

	public string thumbnail;

	public string secure_thumbnail;

	public List<SaleTerms> sale_terms;

	public string video_id;

	public List<string> tags;

	public double? health;

	public string warranty;

	public MeliShipping shipping;

	public string international_delivery_mode;

	public decimal? price;

	public decimal? original_price;

	public decimal? base_price;

	public decimal? quantity;

	public string title;

	public string buying_mode;

	public List<ItemPicuture> pictures;

	public List<MeliAttribute> attributes;

	public List<Variation> variations;

	public string catalog_product_id;

	public bool catalog_listing;

	public List<ItemRelation> item_relations;

	public NetProceeds net_proceeds;

	public bool IsCBT
	{
		get
		{
			if (international_delivery_mode == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C69F388 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376))
			{
				return true;
			}
			if (shipping?.tags != null && shipping.tags.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F97A8FC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a)))
			{
				return true;
			}
			string text = shipping?.logistic_type;
			switch (<PrivateImplementationDetails>.ComputeStringHash(text))
			{
			case 4253581745u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CD53980 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479)))
				{
					goto default;
				}
				return false;
			case 2627737774u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDBAD790 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e)))
				{
					goto default;
				}
				return false;
			case 1971448237u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C02742F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882)))
				{
					goto default;
				}
				return false;
			case 1683458035u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EEC01F7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd)))
				{
					goto default;
				}
				return false;
			case 542584942u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -1996195191 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a21525dd11f740a2800d892a767a04e6)))
				{
					goto default;
				}
				return false;
			case 3679784909u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CDFED38 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479)))
				{
					goto default;
				}
				return false;
			case 2470140894u:
				if (text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3171BF40 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df))
				{
					return false;
				}
				goto default;
			default:
				if (text == null)
				{
					return false;
				}
				return true;
			}
		}
	}

	public int Storage
	{
		get
		{
			if (international_delivery_mode == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43EEF8DB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18))
			{
				return 1;
			}
			if (shipping?.tags != null && shipping.tags.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CDFEEBE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479)))
			{
				return 2;
			}
			string text = shipping?.logistic_type;
			switch (<PrivateImplementationDetails>.ComputeStringHash(text))
			{
			case 4253581745u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FB755E2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff)))
				{
					goto default;
				}
				return 2;
			case 2627737774u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3524612E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45)))
				{
					goto default;
				}
				return 3;
			case 1971448237u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x494B320C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3)))
				{
					goto default;
				}
				return 3;
			case 1683458035u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -793847859 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c)))
				{
					goto default;
				}
				return 3;
			case 542584942u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60885BEB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4)))
				{
					goto default;
				}
				return 3;
			case 3679784909u:
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1851AAFD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233)))
				{
					goto default;
				}
				return 3;
			case 2470140894u:
				if (text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2731FDC1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14))
				{
					return 3;
				}
				goto default;
			default:
				if (text == null)
				{
					return 3;
				}
				return 1;
			}
		}
	}

	public override bool HasError => status.ToInt().Between(400, 600);

	public int Quantity
	{
		get
		{
			if (!available_quantity.HasValue)
			{
				return Math.Max(initial_quantity - sold_quantity.GetValueOrDefault(), 0);
			}
			return available_quantity.Value;
		}
	}

	public bool IsActive => status == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1C2CC982 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21);

	public bool IsFull
	{
		get
		{
			if (shipping == null)
			{
				return false;
			}
			if (shipping.logistic_type == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0CCBDD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c))
			{
				return true;
			}
			return shipping?.tags?.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A20E1D8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed)) == true;
		}
	}

	public int VariationCount => variations?.Count ?? 0;

	public decimal NetProceedsPrice => net_proceeds?.amount ?? 0m;

	public decimal ShipCost => net_proceeds?.ShipCost ?? 0m;

	public decimal SaleFee => net_proceeds?.SaleFee ?? 0m;

	public string FindImgUrl(long? variationid, string sellerSku)
	{
		<>c__DisplayClass74_0 CS$<>8__locals8 = new <>c__DisplayClass74_0();
		CS$<>8__locals8.variationid = variationid;
		CS$<>8__locals8.sellerSku = sellerSku;
		if (variations == null || variations.Count == 0)
		{
			return pictures?.FirstOrDefault()?.secure_url;
		}
		string result = "";
		Variation variation = variations?.FirstOrDefault((Variation i) => i.id == CS$<>8__locals8.variationid);
		if (variation == null && !CS$<>8__locals8.sellerSku.IsNullOrEmpty())
		{
			variation = variations?.FirstOrDefault((Variation i) => i.GetSellerSku() == CS$<>8__locals8.sellerSku);
		}
		if (variation != null)
		{
			<>c__DisplayClass74_1 CS$<>8__locals9 = new <>c__DisplayClass74_1();
			CS$<>8__locals9.picId = variation.picture_ids?.FirstOrDefault();
			result = (CS$<>8__locals9.picId.IsNullOrEmpty() ? null : pictures?.FirstOrDefault((ItemPicuture i) => i.id == CS$<>8__locals9.picId))?.secure_url;
		}
		return result;
	}

	public string GetSellerSku()
	{
		return attributes?.FirstOrDefault((MeliAttribute x) => x.id.EqualIgnore(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1851BC2D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233)))?.value_name;
	}

	public long GetGtinCode()
	{
		return (attributes?.FirstOrDefault((MeliAttribute x) => x.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1705051699 ^ -171238924 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_da7aea68c55141aaa23181fded21b782))?.value_name?.ToLong()).GetValueOrDefault();
	}

	public int? GetWeight()
	{
		if (variations != null && variations.Count > 0)
		{
			int? weight = variations[0].GetWeight();
			if (weight > 0)
			{
				return weight;
			}
		}
		return (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7347654 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x44367B02 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13)))?.GetWeight();
	}

	public int[] GetSize()
	{
		if (variations != null && variations.Count > 0)
		{
			int[] size = variations[0].GetSize();
			if (!size.IsEmpty() && size.Length == 3)
			{
				return size;
			}
		}
		int num = (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -1507594664 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A8C438 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718)))?.GetCM() ?? 0;
		int num2 = (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1851BCF1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FEF8F73 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980)))?.GetCM() ?? 0;
		int num3 = (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3BBD8967 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D2A769 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d)))?.GetCM() ?? 0;
		if (num > 0 && num2 > 0 && num3 > 0)
		{
			return new int[3] { num, num2, num3 };
		}
		return null;
	}

	private MeliAttrValueStruct GetValueStruct(string id)
	{
		<>c__DisplayClass79_0 CS$<>8__locals2 = new <>c__DisplayClass79_0();
		CS$<>8__locals2.id = id;
		return (attributes?.FirstOrDefault((MeliAttribute x) => x.id == CS$<>8__locals2.id))?.ValueStruct;
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
				if (!attribute.value_name.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-813636989) ^ 0x3EF6D1CC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f)))
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

	public MeliItemInfo()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		listing_type_id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE826763 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f);
		base..ctor();
	}
}
