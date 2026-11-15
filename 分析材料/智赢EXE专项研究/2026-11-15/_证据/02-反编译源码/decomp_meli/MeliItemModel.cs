using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Runtime.CompilerServices;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using Enumer;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using ZYing.Meli;
using ZYing.Static;
using ZYing.Web;

namespace ZYing.Crawl;

public class 美客多V2 : 美客多
{
	public class MeliItemModel
	{
		public string id { get; set; }

		public JToken galleryPictures { get; set; }

		public string pictureTemplate { get; set; }

		public string title { get; set; }

		public string subtitle { get; set; }

		public string description { get; set; }

		public string categoryId { get; set; }

		public Price price { get; set; }

		public string shippingCurrencyId { get; set; }

		public decimal shippingCost { get; set; }

		public decimal shippingListCost { get; set; }

		public decimal ShippingFreight
		{
			get
			{
				if (!(shippingListCost > 0m))
				{
					return shippingCost;
				}
				return shippingListCost;
			}
		}

		public string Currency => price?.currency_id;

		public decimal OriginalPrice => price?.original_value ?? 0m;

		public decimal Price => price?.value ?? 0m;

		public Price priceValue { get; set; }

		public long sellerIdFromExperiment { get; set; }

		public JToken outsideVariationsPickers { get; set; }

		public List<VariationsPicker> Pickers
		{
			get
			{
				JToken jToken = outsideVariationsPickers;
				if (jToken == null)
				{
					return null;
				}
				return jToken.ToList<VariationsPicker>();
			}
		}

		public JToken allFeatures { get; set; }

		public List<Allfeature> Features
		{
			get
			{
				JToken jToken = allFeatures;
				return ((jToken != null) ? jToken.ToList<Allfeature>() : null) ?? new List<Allfeature>();
			}
		}

		public string techSpecsCategoryId { get; set; }

		public long techSpecsSellerId { get; set; }

		public List<string> Pictures()
		{
			List<string> list = null;
			switch (galleryPictures?.Type)
			{
			case JTokenType.Object:
			{
				list = new List<string>();
				string text = galleryPictures[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7D933C71 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aa09e57b638a476287c30228f2ba6337)]?.ToString();
				if (!text.IsNullOrEmpty())
				{
					list.Add(string.Format(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x159776B4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120), text));
				}
				break;
			}
			case JTokenType.Array:
				list = galleryPictures?.Select((JToken x) => string.Format(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7D9DFF6F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aa09e57b638a476287c30228f2ba6337), x[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x5369F507 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df)].ToString())).Where((string x) => !x.IsNullOrEmpty())?.ToList() ?? new List<string>();
				break;
			default:
				list = new List<string>();
				break;
			}
			if (outsideVariationsPickers == null)
			{
				return list;
			}
			foreach (VariationsPicker.Product item2 in from x in Pickers.SelectMany((VariationsPicker x) => x.products)
				where !x.PictureId.IsNullOrEmpty()
				select x)
			{
				string item = string.Format(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7659B154 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4), item2.PictureId);
				if (!list.Contains(item))
				{
					list.Add(item);
				}
			}
			return list;
		}

		public int GetSoldQuantity()
		{
			if (subtitle.IsNullOrEmpty())
			{
				return 0;
			}
			string text = Regex.Replace(subtitle, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E033F59 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c), "");
			if (!text.IsNullOrEmpty())
			{
				text = text.Trim().TrimStart('+').Trim();
				return text.ToInt();
			}
			return 0;
		}

		public (NPairCollection pairs, NStockCollection stocks) BuildNStocks()
		{
			NPairCollection nPairCollection = new NPairCollection();
			NStockCollection item = new NStockCollection();
			if (outsideVariationsPickers == null || Pickers.Count == 0)
			{
				return (pairs: nPairCollection, stocks: item);
			}
			foreach (VariationsPicker item2 in Pickers.OrderBy((VariationsPicker x) => x.id))
			{
				nPairCollection.Add(item2.id, (from x in item2.products
					select x.label?.text into x
					where !x.IsNullOrEmpty()
					select x).Distinct().ToArray());
			}
			item = nPairCollection.CreateStocks();
			foreach (NStock item3 in item)
			{
				foreach (VariationsPicker.Product item4 in Pickers.SelectMany((VariationsPicker x) => x.products))
				{
					if (!item4.PictureId.IsNullOrEmpty() && item3.Contains(item4.label?.text ?? ""))
					{
						item3.Pic = string.Format(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D12D3AC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d), item4.PictureId);
						break;
					}
				}
				item3.Num = 100;
			}
			return (pairs: nPairCollection, stocks: item);
		}

		public MeliItemModel()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	public class VariationsPicker
	{
		public class Label
		{
			public string text { get; set; }

			public string color { get; set; }

			public string accessibility_text { get; set; }

			public Label()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public class Selected_Option
		{
			public string text { get; set; }

			public string color { get; set; }

			public Selected_Option()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public class Thumbnail
		{
			public string id { get; set; }

			public Thumbnail()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public class Product
		{
			public string id { get; set; }

			public string attribute_id { get; set; }

			public string preview_id { get; set; }

			public Label label { get; set; }

			public Picture picture { get; set; }

			public string PictureId => picture?.id;

			public string decoration_type { get; set; }

			public Track track { get; set; }

			public bool has_color_picker { get; set; }

			public Product()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public class Picture
		{
			public string id { get; set; }

			public Picture()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public class Track
		{
			public Melidata_Event melidata_event { get; set; }

			public Analytics_Event analytics_event { get; set; }

			public Track()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public class Melidata_Event
		{
			public string path { get; set; }

			public Event_Data event_data { get; set; }

			public Melidata_Event()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public class Event_Data
		{
			public string item_id { get; set; }

			public string picker_id { get; set; }

			public string attribute_id { get; set; }

			public Event_Data()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public class Analytics_Event
		{
			public string section { get; set; }

			public string action { get; set; }

			public string category { get; set; }

			public string label { get; set; }

			public Analytics_Event()
			{
				kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
				OjRa8LWZubDK8CtxBF.HBece8MlL();
				base..ctor();
			}
		}

		public string id { get; set; }

		public bool has_details { get; set; }

		public bool disabled { get; set; }

		public bool is_dropdown { get; set; }

		public Label label { get; set; }

		public Selected_Option selected_option { get; set; }

		public Thumbnail thumbnail { get; set; }

		public List<Product> products { get; set; }

		public string permalink { get; set; }

		public string picker_text_style { get; set; }

		public bool show_divider { get; set; }

		public VariationsPicker()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	public class Price
	{
		public string type { get; set; }

		public decimal value { get; set; }

		public decimal original_value { get; set; }

		public string currency_symbol { get; set; }

		public string currency_id { get; set; }

		public bool billing { get; set; }

		public Price()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	public class Gallerypicture
	{
		public string id { get; set; }

		public string alt { get; set; }

		public int width { get; set; }

		public int height { get; set; }

		public string sanitized_title { get; set; }

		public Gallerypicture()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	public class Allfeature
	{
		public string id { get; set; }

		public string text { get; set; }

		public Allfeature()
		{
			kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
			OjRa8LWZubDK8CtxBF.HBece8MlL();
			base..ctor();
		}
	}

	protected override async Task LoadFromApi(List<string> itemids)
	{
		await Task.CompletedTask;
	}

	protected override async Task<SpData> GetSpData()
	{
		await 检查采集数量();
		MeliUrlInfo urlInfo = MeliUrlInfo.Parse(Url);
		AppConfig cfg = JsonConfig<AppConfig>.Current;
		SpData sd = new SpData
		{
			SiteID = 8,
			Lan = Enumer.Meli.GetLan(urlInfo.Site),
			Cur = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1599B5A0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120),
			Area = urlInfo.Site,
			Source = urlInfo.ItemLink,
			ReplacePair = cfg.ReplacePair
		};
		if (cfg.NoCatalogItem && !urlInfo.ProductId.IsNullOrEmpty())
		{
			sd.Error = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x6A705991 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb);
			return sd;
		}
		try
		{
			await FillItemId(urlInfo);
			HttpClient httpClient = HttpClientEx.Create(90);
			httpClient.DefaultRequestHeaders.TryAddWithoutValidation(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x273FA447 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -608419186 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa));
			httpClient.DefaultRequestHeaders.TryAddWithoutValidation(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A2F9CB4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8F3C3C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6));
			string text = (cfg.UsePrivateServer ? kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x220DCF95 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2) : kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE7EC61 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04));
			MeliItemModel model = JsonConvert.DeserializeObject<MeliItemModel>(await Sender(httpClient, text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x687875C1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e92d645a8d04335b6bb7b1de9e0a791) + urlInfo.ItemId));
			await FillModel(sd, model);
			UploadPayload(model).GetAwaiter();
		}
		catch (Exception ex)
		{
			sd.Error = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3D1C87A8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a) + ex.Message;
		}
		return sd;
	}

	private async Task<string> Sender(HttpClient client, string url)
	{
		string result = default;
		object obj;
		int num;
		try
		{
			result = await client.GetStringAsync(url);
			return result;
		}
		catch (Exception ex)
		{
			obj = ex;
			num = 1;
		}
		if (num != 1)
		{
			return result;
		}
		_ = (Exception)obj;
		return await client.GetStringAsync(url);
	}

	private async Task FillModel(SpData sd, MeliItemModel model)
	{
		if (await SellerExists(model.techSpecsSellerId))
		{
			sd.Error = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x273F449B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14);
			return;
		}
		AppConfig current = JsonConfig<AppConfig>.Current;
		sd.ReplacePair = current.ReplacePair;
		sd.Buys = model.GetSoldQuantity();
		if (current.OnlyBuy)
		{
			if (sd.Buys == 0)
			{
				sd.Error = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x23683DD6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8);
				return;
			}
			if (sd.Buys < current.GteSale)
			{
				DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(10, 1);
				defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EE92003 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd));
				defaultInterpolatedStringHandler.AppendFormatted(current.GteSale);
				defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x608D7A61 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4));
				sd.Error = defaultInterpolatedStringHandler.ToStringAndClear();
				return;
			}
		}
		decimal num = ((model.OriginalPrice > 0m) ? model.OriginalPrice : model.Price);
		decimal money = (current.DiscountPrice ? model.Price : num);
		string curByKey = Enumer.Meli.GetCurByKey(sd.Area);
		sd.Cost = ZExchange.Convert(money, curByKey, sd.Cur);
		if (current.MinPirce > (decimal?)0m || current.MaxPirce > (decimal?)0m)
		{
			decimal cost = sd.Cost;
			decimal? minPirce = current.MinPirce;
			if (!((decimal?)cost < minPirce))
			{
				decimal cost2 = sd.Cost;
				minPirce = current.MaxPirce;
				if (!((decimal?)cost2 > minPirce))
				{
					goto IL_03c7;
				}
			}
			DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(18, 3);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x317F0028 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df));
			defaultInterpolatedStringHandler.AppendFormatted(sd.Cost);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x220DBBC7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2));
			defaultInterpolatedStringHandler.AppendFormatted(current.MinPirce);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D7225F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d));
			defaultInterpolatedStringHandler.AppendFormatted(current.MaxPirce);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1890D8DA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba));
			sd.Error = defaultInterpolatedStringHandler.ToStringAndClear();
			return;
		}
		goto IL_03c7;
		IL_03c7:
		sd.Shop = (sd.Seller = model.techSpecsSellerId.ToString());
		sd.Title = model.title;
		sd.Intro = model.description;
		sd.Pics = model.Pictures();
		sd.WAttr = new NAttrCollection
		{
			Site = sd.Area,
			KindID = model.categoryId.Substring(3),
			Attrs = new List<NAttr>()
		};
		foreach (Allfeature feature in model.Features)
		{
			sd.WAttr.Attrs.Add(new NAttr
			{
				Name = feature.id,
				Value = feature.text
			});
		}
		(NPairCollection pairs, NStockCollection stocks) tuple = model.BuildNStocks();
		NPairCollection item = tuple.pairs;
		NStockCollection item2 = tuple.stocks;
		sd.Pairs = item;
		sd.Stocks = item2;
		sd.Stocks.FillPics(sd.Pics);
		await FillFreight(sd, model.id, model.shippingCurrencyId, model.shippingCost, model.shippingListCost);
	}

	private async Task UploadPayload(MeliItemModel model)
	{
		int soldQuantity = model.GetSoldQuantity();
		if (model.id.IsNullOrEmpty() || soldQuantity <= 0)
		{
			return;
		}
		string text = model.id.Substring(0, 3);
		int site;
		if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C6354D2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)))
		{
			if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x23322C9E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1)))
			{
				if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1D2A3C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d)))
				{
					if (!(text == kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x310C868 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339)))
					{
						return;
					}
					site = 4;
				}
				else
				{
					site = 3;
				}
			}
			else
			{
				site = 2;
			}
		}
		else
		{
			site = 1;
		}
		PayloadInfo payloadInfo = new PayloadInfo
		{
			Site = site,
			Url = Enumer.Meli.GetItemLink(model.id)
		};
		payloadInfo.Items.Add(new PayloadInfo.Item
		{
			ItemId = model.id,
			OgiginalPrice = model.OriginalPrice,
			Price = model.Price,
			Sold = soldQuantity,
			SoldRange = (soldQuantity > 5)
		});
		await UploadPayload(payloadInfo);
	}

	private async Task FillItemId(MeliUrlInfo info)
	{
		if (!info.ProductId.IsNullOrEmpty() && info.ItemId.IsNullOrEmpty())
		{
			string text = (await (await MeliApiBuilder.Create()).GetProductItems(info.ProductId)).results?.FirstOrDefault().item_id;
			if (text.IsNullOrEmpty())
			{
				throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FB8D31E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff) + info.ProductId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46F51759 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51));
			}
			info.ItemId = text;
		}
	}

	public 美客多V2()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}
}
