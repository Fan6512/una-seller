using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Web;

namespace ZYing.Meli.Model;

public class Variation
{
	[CompilerGenerated]
	private sealed class <>c__DisplayClass14_0
	{
		public string id;

		public <>c__DisplayClass14_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <SetAttr>b__0(MeliAttribute x)
		{
			return x.id == id;
		}
	}

	[CompilerGenerated]
	private sealed class <>c__DisplayClass20_0
	{
		public string id;

		public <>c__DisplayClass20_0()
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

	public long? id;

	public decimal? price;

	public int? available_quantity;

	public int? sold_quantity;

	public string inventory_id;

	public List<MeliAttribute> attribute_combinations;

	public List<MeliAttribute> attributes;

	public List<string> picture_ids;

	[JsonIgnore]
	public string VariKey;

	public void SetWeight(int weight, string unit = "g")
	{
		string text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE826785 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f);
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(1, 2);
		defaultInterpolatedStringHandler.AppendFormatted(weight);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2366A150 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8));
		defaultInterpolatedStringHandler.AppendFormatted(unit);
		SetAttr(text, defaultInterpolatedStringHandler.ToStringAndClear());
	}

	public void SetSize(int[] size, string unit = "cm")
	{
		string text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x47880079 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c);
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(1, 2);
		defaultInterpolatedStringHandler.AppendFormatted(size[0]);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE89CCB1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f));
		defaultInterpolatedStringHandler.AppendFormatted(unit);
		SetAttr(text, defaultInterpolatedStringHandler.ToStringAndClear());
		string text2 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x15923C64 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120);
		defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(1, 2);
		defaultInterpolatedStringHandler.AppendFormatted(size[1]);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4783AB7D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c));
		defaultInterpolatedStringHandler.AppendFormatted(unit);
		SetAttr(text2, defaultInterpolatedStringHandler.ToStringAndClear());
		string text3 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -1784732047 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e);
		defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(1, 2);
		defaultInterpolatedStringHandler.AppendFormatted(size[2]);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FE42439 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980));
		defaultInterpolatedStringHandler.AppendFormatted(unit);
		SetAttr(text3, defaultInterpolatedStringHandler.ToStringAndClear());
	}

	public void SetWeightSize(int weight, int[] size)
	{
		if (!size.IsEmpty())
		{
			string text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDBAC19C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e);
			DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(2, 1);
			defaultInterpolatedStringHandler.AppendFormatted(weight);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FEF8FD9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980));
			SetAttr(text, defaultInterpolatedStringHandler.ToStringAndClear());
			string text2 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A8C458 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718);
			defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(3, 1);
			defaultInterpolatedStringHandler.AppendFormatted(size[0]);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1895E95C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba));
			SetAttr(text2, defaultInterpolatedStringHandler.ToStringAndClear());
			string text3 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E55CB97 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4);
			defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(3, 1);
			defaultInterpolatedStringHandler.AppendFormatted(size[1]);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43EEEDE7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18));
			SetAttr(text3, defaultInterpolatedStringHandler.ToStringAndClear());
			string text4 = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3D1A5AA2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a);
			defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(3, 1);
			defaultInterpolatedStringHandler.AppendFormatted(size[2]);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE2A853 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04));
			SetAttr(text4, defaultInterpolatedStringHandler.ToStringAndClear());
		}
	}

	public void SetSellerSku(string sku)
	{
		SetAttr(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1799A0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d), sku);
	}

	public void SetGtinCode(string code)
	{
		SetAttr(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7D93723B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aa09e57b638a476287c30228f2ba6337), code);
		attributes.RemoveAll((MeliAttribute x) => x.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3137BED7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1));
	}

	public void SetAttr(string id, string valueName)
	{
		<>c__DisplayClass14_0 CS$<>8__locals3 = new <>c__DisplayClass14_0();
		CS$<>8__locals3.id = id;
		if (attributes == null)
		{
			attributes = new List<MeliAttribute>();
		}
		MeliAttribute meliAttribute = attributes.FirstOrDefault((MeliAttribute x) => x.id == CS$<>8__locals3.id);
		if (meliAttribute != null)
		{
			meliAttribute.value_name = valueName;
			return;
		}
		attributes.Add(new MeliAttribute
		{
			id = CS$<>8__locals3.id,
			value_name = valueName
		});
	}

	public string GetSellerSku()
	{
		return attributes?.FirstOrDefault((MeliAttribute x) => x.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11039026 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315))?.value_name;
	}

	public long GetGtinCode()
	{
		return (attributes?.FirstOrDefault((MeliAttribute x) => x.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x76F2C4A3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45))?.value_name?.ToLong()).GetValueOrDefault();
	}

	public string GetSizeChartRowId()
	{
		return attributes?.FirstOrDefault((MeliAttribute x) => x.id == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x731F3AA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3))?.value_name;
	}

	public int? GetWeight()
	{
		return (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x15923C04 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3D1A5A0E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a)))?.GetWeight();
	}

	public int[] GetSize()
	{
		int num = (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236D0A54 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x226A7935 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd)))?.GetCM() ?? 0;
		int num2 = (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A20F452 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x22088AF5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2)))?.GetCM() ?? 0;
		int num3 = (GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA80DA6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22)) ?? GetValueStruct(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x44367B8E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13)))?.GetCM() ?? 0;
		if (num > 0 && num2 > 0 && num3 > 0)
		{
			return new int[3] { num, num2, num3 };
		}
		return null;
	}

	private MeliAttrValueStruct GetValueStruct(string id)
	{
		<>c__DisplayClass20_0 CS$<>8__locals2 = new <>c__DisplayClass20_0();
		CS$<>8__locals2.id = id;
		return (attributes?.FirstOrDefault((MeliAttribute x) => x.id == CS$<>8__locals2.id))?.ValueStruct;
	}

	public string GetVariationKey()
	{
		List<MeliAttribute> list = attribute_combinations;
		if (list == null)
		{
			return null;
		}
		return list.Select((MeliAttribute x) => x.value_name).Join(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2BABEE44 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e92d645a8d04335b6bb7b1de9e0a791));
	}

	public Variation()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}
}
