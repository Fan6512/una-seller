using System.Collections.Generic;
using System.Linq;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Data;
using ZYing.Meli.Model;

namespace ZYing.Meli;

public class MeliUpVariants : List<MeliUpVariantItem>
{
	public Record ToRecord()
	{
		Record record = new Record();
		if (Count <= 1)
		{
			MeliUpVariantItem meliUpVariantItem = this.FirstOrDefault();
			record.AddRow();
			record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x15991B48 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120)).Set(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1895EF76 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba));
			record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7C689F94 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_db457e1da79640b7816cb6ccbfc1d724)).Set(meliUpVariantItem?.seller_sku);
			record.Columns.AddLong(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1C2D8332 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21)).Set(meliUpVariantItem?.code);
			record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F458F12 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb)).Set(meliUpVariantItem?.item_id);
			record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A0DAB53 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0bd1e5ac52394aa1bf047220642aab8c)).Set(meliUpVariantItem?.user_product_id);
			record.Columns.AddInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x73FB74A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3)).Set(meliUpVariantItem?.status);
			record.Columns.AddInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -585179829 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a)).Set(meliUpVariantItem?.available_quantity);
			record.Columns.AddInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F97BA5C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a)).Set(meliUpVariantItem?.sold_quantity);
			record.Columns.AddDecimal(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2366EDC0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8)).Set(meliUpVariantItem?.price);
			record.Columns.AddDecimal(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE8260ED ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f)).Set(meliUpVariantItem?.net_proceeds);
			record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -1549707713 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e)).Set(meliUpVariantItem?.ErrorMsg);
		}
		else
		{
			using Enumerator enumerator = GetEnumerator();
			while (enumerator.MoveNext())
			{
				MeliUpVariantItem current = enumerator.Current;
				record.AddRow();
				ItemPicuture itemPicuture = current.pictures.FirstOrDefault((ItemPicuture x) => !x.source.IsNullOrEmpty() || !x.id.IsNullOrEmpty());
				if (itemPicuture != null)
				{
					record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1312DDA4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82a9ebeebdaa41eeb6924eb98e4d50c1)).Set(itemPicuture.Url_2X);
				}
				record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1CBE50 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d)).Set(current.key);
				record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1316A612 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82a9ebeebdaa41eeb6924eb98e4d50c1)).Set(current?.seller_sku);
				record.Columns.AddLong(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x612639E5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f8d88042fa6e4eae8f42d355fd0508fd)).Set(current?.code);
				record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x18952EE8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba)).Set(current?.item_id);
				record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C69E168 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)).Set(current?.user_product_id);
				record.Columns.AddInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C622776 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)).Set(current?.status);
				record.Columns.AddInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22088DA9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2)).Set(current?.available_quantity);
				record.Columns.AddInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FEF8847 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980)).Set(current?.sold_quantity);
				record.Columns.AddDecimal(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F9C5AB2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a)).Set(current?.price);
				record.Columns.AddDecimal(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B993EF9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c)).Set(current?.net_proceeds);
				record.Columns.AddString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3171B99C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df)).Set(current?.ErrorMsg);
			}
		}
		return record;
	}

	public MeliUpVariants()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}
}
