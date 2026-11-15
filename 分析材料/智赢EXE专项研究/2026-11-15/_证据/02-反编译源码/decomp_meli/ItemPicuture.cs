using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;

namespace ZYing.Meli.Model;

public class ItemPicuture
{
	public string source;

	public string id;

	public string url;

	public string secure_url;

	public string size;

	public string max_size;

	public string Url_2X
	{
		get
		{
			if (!id.IsNullOrEmpty())
			{
				return kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CDFEDD2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479) + id + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -1722817595 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6);
			}
			return (secure_url ?? source ?? url)?.Replace(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x5D17842E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E061FCD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c));
		}
	}

	public ItemPicuture()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}
}
