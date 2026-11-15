using System;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Web;

namespace ZYing;

[JsonLower]
public class AppConfig : JsonConfig<AppConfig>
{
	public class MeliSubmitPara
	{
		public bool Select_Weight;

		public bool Add_Freight_Enable;

		public string[] Except_Sites;

		public string[] Include_Lans;

		public string[] AbbrNames { get; set; }

		public bool NotLoadCates { get; set; }

		public MeliSubmitPara()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			Except_Sites = new string[2]
			{
				$"{8}",
				$"{1}"
			};
			Include_Lans = new string[1] { kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E5E60BD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4) };
			base..ctor();
		}
	}

	public bool HK;

	public bool DelSame;

	public bool ChinaUp;

	public string ImageDir;

	public string Url;

	public string OzonUrl;

	public bool UseMeliServer;

	public bool UsePrivateServer;

	public bool NoCatalogItem;

	public bool NoOffical;

	public bool NoDepot;

	public bool NoLocal;

	public bool OnlyBuy;

	public bool OnlyChina;

	public bool OnlyTitle;

	public bool DiscountPrice;

	public bool ReplacePair;

	public string Parts;

	public string MeliParts;

	public string YandexParts;

	public string TaskParts;

	public string TaskMeliParts;

	public string TaskYandexParts;

	public string Prompt;

	public bool Exclude;

	public int GteSale;

	public string Stock;

	public decimal? MinPirce;

	public decimal? MaxPirce;

	public int OrderSize;

	public int SaleSize;

	public bool IgnoreRepeatVariant;

	public string ProxyEndpoint;

	public string ProxyUsername;

	public string ProxyPassword;

	public string[] Provinces;

	public string[] AotArea;

	public bool LastLoginIsZeyuan;

	public int AmzFollowSpiderImport_PriceFrom;

	public int AmzFollowSpiderImport_SkuFormat;

	public decimal AmzFollowSpiderImport_MaxPriceRate;

	public decimal AmzFollowSpiderImport_MinPriceRate;

	public int AmzFollowSpiderImport_Num;

	public int AmzFollowSpiderImport_Delay;

	public int AmzFollowSpiderImport_PriceId;

	public string ZFollow_Calc_Config;

	public int FollowSetAidOldAccPolicy;

	public bool PrintPickLabel;

	public DateTime AlertTime;

	public bool NotAlert;

	public bool NotAlertForCharge;

	public string ZtoDevice;

	public string ExportOrderColumns;

	public int OzonBarcodeLabelW;

	public int OzonBarcodeLabelH;

	public MeliSubmitPara MeliPara;

	public AppConfig()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		ImageDir = "";
		Url = "";
		OzonUrl = "";
		ReplacePair = true;
		ProxyEndpoint = string.Empty;
		ProxyUsername = string.Empty;
		ProxyPassword = string.Empty;
		AmzFollowSpiderImport_PriceFrom = 1;
		AmzFollowSpiderImport_SkuFormat = 1;
		AmzFollowSpiderImport_Num = 20;
		ZFollow_Calc_Config = string.Empty;
		FollowSetAidOldAccPolicy = 1;
		ZtoDevice = "";
		OzonBarcodeLabelW = 100;
		OzonBarcodeLabelH = 50;
		base..ctor();
	}
}
