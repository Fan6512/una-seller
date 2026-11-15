using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Meli.Model;

namespace ZYing.Meli;

public class MeliAttrDto : MeliIdName
{
	private static string[] DefaultStrs;

	public bool Required { get; set; }

	public bool VariationAttr { get; set; }

	public bool AllowVariation { get; set; }

	public bool Hidden { get; set; }

	public bool ReadOnly { get; set; }

	public string ValueType { get; set; }

	public int ValueMaxLength { get; set; }

	public bool Multi { get; set; }

	public string DefaultUnit { get; set; }

	public List<string> Units { get; set; }

	public bool AllowUnit
	{
		get
		{
			if (Units != null)
			{
				return Units.Count > 0;
			}
			return false;
		}
	}

	public List<MeliIdName> Values { get; set; }

	public IEnumerable<int> ValueIds => Values?.Select((MeliIdName x) => x.Id.ToInt());

	public override string ToString()
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(2, 3);
		defaultInterpolatedStringHandler.AppendFormatted(Name);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0DA23D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c));
		defaultInterpolatedStringHandler.AppendFormatted(NameZh);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE89CCB1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f));
		defaultInterpolatedStringHandler.AppendFormatted(ValueType);
		return defaultInterpolatedStringHandler.ToStringAndClear();
	}

	public void FillDefaultValue(MeliAttribute attr)
	{
		string valueType = ValueType;
		if (!(valueType == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EEC137F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd)))
		{
			if (!(valueType == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C0266C9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882)))
			{
				if (!(valueType == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1319F978 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82a9ebeebdaa41eeb6924eb98e4d50c1)))
				{
					if (!(valueType == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43EEE9D9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18)))
					{
						if (valueType == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x765CFF18 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4))
						{
							attr.value_id = Values?.FirstOrDefault()?.Id;
						}
					}
					else
					{
						attr.value_name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3D1A5F20 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a) + (DefaultUnit ?? Units.FirstOrDefault());
					}
				}
				else
				{
					attr.value_name = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -1093541192 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c);
				}
			}
			else
			{
				attr.value_id = ValueIds?.FirstOrDefault().ToString();
			}
		}
		else
		{
			attr.value_name = DefaultStrs[Rnd.Next(0, DefaultStrs.Length)];
		}
	}

	public MeliAttrDto()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}

	static MeliAttrDto()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		DefaultStrs = new string[2]
		{
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE11412D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60884885 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4)
		};
	}
}
