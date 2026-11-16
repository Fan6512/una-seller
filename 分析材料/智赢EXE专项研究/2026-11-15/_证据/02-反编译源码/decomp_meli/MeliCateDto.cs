using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Meli.Model;

namespace ZYing.Meli;

public class MeliCateDto
{
	public string Site { get; set; }

	public int Id { get; set; }

	public string CateId => Site + Id;

	public string Name { get; set; }

	public string NameZh { get; set; }

	public string FullName { get; set; }

	public string FullZh { get; set; }

	public bool IsLeaf { get; set; }

	public int ReferCbtId { get; set; }

	public string Domain { get; set; }

	public int MaxTitleLength { get; set; }

	public bool AllowVariations { get; set; }

	public int MaxVariations { get; set; }

	public int MaxPictures { get; set; }

	public int MaxPicturesVar { get; set; }

	public List<MeliAttrDto> Attributes { get; set; }

	public DateTime UpdateAt { get; set; }

	public IEnumerable<MeliAttrDto> Variations => from x in Attributes
		where x.AllowVariation && !x.Hidden
		orderby x.Required descending
		select x;

	public IEnumerable<MeliAttrDto> VariationAttrs => Attributes.Where((MeliAttrDto x) => x.VariationAttr && x.Id.NotIn(MeliConsts.AttrNames.CommonAttrs));

	public IEnumerable<MeliAttrDto> ItemAttrs => Attributes.Where((MeliAttrDto x) => !x.AllowVariation && x.Id.NotIn(MeliConsts.AttrNames.CommonAttrs));

	public override string ToString()
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(8, 3);
		defaultInterpolatedStringHandler.AppendFormatted(CateId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDB16AA8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e));
		defaultInterpolatedStringHandler.AppendFormatted(NameZh);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FEF8BAB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980));
		defaultInterpolatedStringHandler.AppendFormatted(Attributes.Count);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -327439061 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a));
		return defaultInterpolatedStringHandler.ToStringAndClear();
	}

	public MeliCateDto()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		MaxTitleLength = 60;
		MaxVariations = 100;
		MaxPictures = 12;
		MaxPicturesVar = 10;
		Attributes = new List<MeliAttrDto>();
		base..ctor();
	}
}
