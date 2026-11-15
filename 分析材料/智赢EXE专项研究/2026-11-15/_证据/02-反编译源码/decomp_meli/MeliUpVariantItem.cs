using System;
using System.Collections.Generic;
using System.Linq;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using Newtonsoft.Json;
using ZYing.Meli.Model;
using ZYing.Web;

namespace ZYing.Meli;

public class MeliUpVariantItem
{
	public string item_id;

	public string cbt_id;

	public string user_product_id;

	public string catalog_product_id;

	public string siteless_up_id;

	public long? family_id;

	public decimal? price;

	public decimal? original_price;

	public decimal? net_proceeds;

	public decimal? shipping_cost;

	public decimal? sale_fee;

	[ZYing.Web.JsonIgnore]
	[Newtonsoft.Json.JsonIgnore]
	public List<MeliAttribute> VariantAttrs;

	public string key;

	public string seller_sku;

	public long? code;

	public int? available_quantity;

	public int? sold_quantity;

	public int? status;

	public List<ItemPicuture> pictures;

	public DateTime date_created;

	public string ErrorMsg;

	public string GetVariantKey()
	{
		List<MeliAttribute> variantAttrs = VariantAttrs;
		if (variantAttrs == null)
		{
			return null;
		}
		IEnumerable<string> enumerable = variantAttrs.Select((MeliAttribute x) => x.value_name);
		if (enumerable == null)
		{
			return null;
		}
		return enumerable.Join(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C09C8A9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882));
	}

	public MeliUpVariantItem()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		pictures = new List<ItemPicuture>();
		base..ctor();
	}
}
