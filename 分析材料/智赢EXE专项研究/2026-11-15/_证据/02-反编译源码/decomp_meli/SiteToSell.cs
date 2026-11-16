using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Web;

namespace ZYing.Meli.Models.UserProducts;

public class SiteToSell
{
	public string site_id;

	public string logistic_type;

	public decimal? price;

	public decimal? net_proceeds;

	public string listing_type_id;

	public string title;

	public ProductDescription description;

	public string Id => site_id + logistic_type;

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

	public SiteToSell()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		listing_type_id = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794B75A0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6);
		base..ctor();
	}
}
