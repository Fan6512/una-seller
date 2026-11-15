using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Web;

namespace ZYing.Meli.Model;

public class MeliAttribute
{
	[CompilerGenerated]
	private sealed class <>c__DisplayClass15_0
	{
		public NAttr attr;

		public <>c__DisplayClass15_0()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}

		internal bool <FromAttrJson>b__0(MeliAttribute x)
		{
			return x.id.Contains(attr.NameId);
		}
	}

	public string id;

	public string name;

	public string value_id;

	public string value_name;

	public string value_type;

	public List<MeliCateAttr_Value> values;

	public MeliAttrValueStruct value_struct;

	public List<MeliCateAttr_Value> allowed_units;

	public MeliCateAttr_Tag tags;

	public MeliAttrValueStruct ValueStruct
	{
		get
		{
			if (value_struct == null)
			{
				return values?.FirstOrDefault()?.@struct;
			}
			return value_struct;
		}
	}

	[JsonIgnore]
	public bool AllowVariations
	{
		get
		{
			return tags?.allow_variations == true;
		}
		set
		{
			if (tags == null)
			{
				tags = new MeliCateAttr_Tag();
			}
			tags.allow_variations = value;
		}
	}

	public void AddValue(string id, string name)
	{
		if (values == null)
		{
			values = new List<MeliCateAttr_Value>();
		}
		MeliCateAttr_Value item = new MeliCateAttr_Value
		{
			id = id,
			name = name
		};
		values.Add(item);
	}

	public static List<MeliAttribute> FromAttrJson(string json)
	{
		List<NAttr> list = Json.ReadList<NAttr>(json);
		List<MeliAttribute> list2 = new List<MeliAttribute>();
		using List<NAttr>.Enumerator enumerator = list.GetEnumerator();
		while (enumerator.MoveNext())
		{
			<>c__DisplayClass15_0 CS$<>8__locals8 = new <>c__DisplayClass15_0();
			CS$<>8__locals8.attr = enumerator.Current;
			MeliAttribute meliAttribute = new MeliAttribute
			{
				id = CS$<>8__locals8.attr.NameId,
				name = CS$<>8__locals8.attr.Name
			};
			string text = ((CS$<>8__locals8.attr.ValueId > 0) ? CS$<>8__locals8.attr.ValueId.ToString() : null);
			if (list2.Contains((MeliAttribute x) => x.id.Contains(CS$<>8__locals8.attr.NameId)))
			{
				meliAttribute.AddValue(text, CS$<>8__locals8.attr.Value);
				continue;
			}
			meliAttribute.value_id = text;
			meliAttribute.value_name = CS$<>8__locals8.attr.Value;
			list2.Add(meliAttribute);
		}
		return list2;
	}

	public int GetNumber()
	{
		int? num;
		if (value_struct == null)
		{
			num = values?.FirstOrDefault()?.GetNumber();
		}
		else
		{
			MeliAttrValueStruct meliAttrValueStruct = value_struct;
			if (meliAttrValueStruct == null)
			{
				num = null;
			}
			else
			{
				ref double? number = ref meliAttrValueStruct.number;
				num = (number.HasValue ? new int?(number.GetValueOrDefault().ToInt()) : ((int?)null));
			}
		}
		int? num2 = num;
		return num2.GetValueOrDefault();
	}

	public int? GetNumberMatchUnit(string unit)
	{
		if (value_name.IsNullOrEmpty())
		{
			return null;
		}
		string[] array = Split.String(value_name, ' ');
		if (!((array.Length > 1) ? array[1] : "").Equals(unit ?? "", StringComparison.OrdinalIgnoreCase))
		{
			return null;
		}
		return array[0].ToInt();
	}

	public override string ToString()
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(13, 3);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x23389FD4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1));
		defaultInterpolatedStringHandler.AppendFormatted(id);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236D0B42 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8));
		defaultInterpolatedStringHandler.AppendFormatted(value_id);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43E5699B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18));
		defaultInterpolatedStringHandler.AppendFormatted(value_name);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189EA790 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba));
		return defaultInterpolatedStringHandler.ToStringAndClear();
	}

	public MeliAttribute()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}
}
