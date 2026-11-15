using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net.Http;
using System.Runtime.CompilerServices;
using System.Runtime.ExceptionServices;
using System.Threading.Tasks;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using ZYing.Data;
using ZYing.Logging;
using ZYing.Meli.Model;
using ZYing.Meli.Models.UserProducts;
using ZYing.Web;

namespace ZYing.Meli;

public class MeliApi
{
	private static Dictionary<string, string> AuthUrls;

	public const string ApiHost = "https://api.mercadolibre.com";

	private static HttpClient _h;

	private string _token;

	private DateTime _expire;

	private string _refresh;

	private bool _isCbt;

	public int Aid { get; }

	public int GroupId { get; }

	public int UserId { get; }

	public long ShopId { get; }

	public string SiteId { get; }

	public long SiteUserId { get; }

	public string Abbr { get; }

	public bool IsCbt => _isCbt;

	public bool IsFull { get; }

	public string AppId { get; }

	public static string GetAuthUrl(string siteId, string state = null)
	{
		MeliAppInfo appInfo = MeliAppInfo.GetAppInfo(siteId);
		if (!AuthUrls.TryGetValue(siteId, out var value))
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE119685 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152) + siteId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x765C292E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4));
		}
		string value2 = value;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(65, 4);
		defaultInterpolatedStringHandler.AppendFormatted(value2);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3D1A88F6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a));
		defaultInterpolatedStringHandler.AppendFormatted(appInfo.AppId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1272AEC3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff));
		defaultInterpolatedStringHandler.AppendFormatted(Uri.EscapeDataString(appInfo.RedirectUrl));
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x8668DEC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb));
		defaultInterpolatedStringHandler.AppendFormatted(Uri.EscapeDataString(state));
		return defaultInterpolatedStringHandler.ToStringAndClear();
	}

	public static async Task<TokenInfo> GetToken(string code, string siteId)
	{
		MeliAppInfo appInfo = MeliAppInfo.GetAppInfo(siteId);
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(116, 4);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31C0014 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339));
		defaultInterpolatedStringHandler.AppendFormatted(appInfo.AppId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F876DA4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6));
		defaultInterpolatedStringHandler.AppendFormatted(appInfo.AppSecret);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E016DC7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c));
		defaultInterpolatedStringHandler.AppendFormatted(code);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F81FD9C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6));
		defaultInterpolatedStringHandler.AppendFormatted(appInfo.RedirectUrl);
		string requestUri = defaultInterpolatedStringHandler.ToStringAndClear();
		using HttpRequestMessage req = new HttpRequestMessage(HttpMethod.Post, requestUri);
		using HttpResponseMessage rsp = await _h.SendAsync(req);
		string text = await rsp.Content.ReadAsStringAsync();
		if (!rsp.IsSuccessStatusCode)
		{
			MeliErrorParser.Instance.TryThrow((int)rsp.StatusCode, text);
		}
		return Json.Read<TokenInfo>(text);
	}

	public static async Task<TokenInfo> RefreshToken(string refreshToken, string accessToken)
	{
		string text = ParseAppId(accessToken);
		MeliAppInfo byAppId = MeliAppInfo.GetByAppId(text);
		if (byAppId == null)
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CAE783C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22) + text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8C0D52 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6));
		}
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(106, 3);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x7FDF58D5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152));
		defaultInterpolatedStringHandler.AppendFormatted(byAppId.AppId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6FB082E4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a89dc9f0245e4a528323c23b24f201ff));
		defaultInterpolatedStringHandler.AppendFormatted(byAppId.AppSecret);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1104F13C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315));
		defaultInterpolatedStringHandler.AppendFormatted(refreshToken);
		string requestUri = defaultInterpolatedStringHandler.ToStringAndClear();
		using HttpRequestMessage req = new HttpRequestMessage(HttpMethod.Post, requestUri);
		using HttpResponseMessage rsp = await _h.SendAsync(req);
		string text2 = await rsp.Content.ReadAsStringAsync();
		if (!rsp.IsSuccessStatusCode)
		{
			MeliErrorParser.Instance.TryThrow((int)rsp.StatusCode, text2);
		}
		return Json.Read<TokenInfo>(text2);
	}

	public static string ParseAppId(string accessToken)
	{
		int num = accessToken.IndexOf('-');
		string text = accessToken.Substring(num + 1);
		return text.Substring(0, text.IndexOf('-'));
	}

	public async Task<SerachCatalogsResponse> SerachCatalogs(CatalogQuery query)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(21, 2);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3134421B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1));
		defaultInterpolatedStringHandler.AppendFormatted(ShopId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6F985167 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_da7aea68c55141aaa23181fded21b782));
		defaultInterpolatedStringHandler.AppendFormatted(query.QueryString());
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		return Deserialize<SerachCatalogsResponse>(await SendAsync(url));
	}

	public async Task<CatalogEligibilityResponse> CatalogListingEligibility(string itemid)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7C61A58A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_db457e1da79640b7816cb6ccbfc1d724) + itemid + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D4C67D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d);
		return Deserialize<CatalogEligibilityResponse>(await SendAsync(url));
	}

	public async Task<string> CatalogListingEligibility(List<string> itemids)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D11F8E8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d) + itemids.Join(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E5E7511 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4));
		return await SendAsync(url);
	}

	public async Task<SearchProductResponse> SearchProducts(SearchProductQuery query)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE8411B5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f) + query.QueryString();
		return Deserialize<SearchProductResponse>(await SendAsync(url));
	}

	public async Task<CatalogProduct> GetCatalogProduct(string productId)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x220F8C6B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2) + productId;
		return Deserialize<CatalogProduct>(await SendAsync(url));
	}

	public async Task<MeliItemInfo> PublishCatalog(string itemid, string catalog_product_id, long? variation_id = null)
	{
		PublishCatalogListing o = new PublishCatalogListing
		{
			item_id = itemid,
			catalog_product_id = catalog_product_id,
			variation_id = variation_id
		};
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1705051699 ^ -432060155 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_db457e1da79640b7816cb6ccbfc1d724);
		return Deserialize<MeliItemInfo>(await SendAsync(url, HttpMethod.Post, o.ToJson()));
	}

	public async Task<PriceToWinResult> PriceToWin(string itemid)
	{
		if (itemid.IsNullOrEmpty())
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86B5056 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb));
		}
		string value = itemid.Substring(0, 3);
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(51, 2);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CD985D8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479));
		defaultInterpolatedStringHandler.AppendFormatted(itemid);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x35EE629C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4d8504ccce2f465ba969392f5e065f21));
		defaultInterpolatedStringHandler.AppendFormatted(value);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7B9130B2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4d8504ccce2f465ba969392f5e065f21));
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		return Deserialize<PriceToWinResult>(await SendAsync(url));
	}

	public async Task<ProductItemsResult> GetProductItems(string productId)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x352371C2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45) + productId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A46A0F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd);
		return Deserialize<ProductItemsResult>(await SendAsync(url));
	}

	public async Task<List<MeliCategory>> GetRootCategories(string siteId)
	{
		return Deserialize<List<MeliCategory>>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x317C4C30 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df) + siteId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1857C645 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233)));
	}

	public async Task<List<MeliCategory>> GetAllCategories(string site)
	{
		return Deserialize<MeliAllCategory>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x127F560F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff) + site + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EEA6AF1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd))).BuildCateTree();
	}

	public async Task<MeliCategory> GetCategory(string cateId)
	{
		return Deserialize<MeliCategory>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CAE7070 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22) + cateId));
	}

	public async Task<List<MeliAttr>> GetCateAttrs(string cateId, bool throwError = true)
	{
		return Deserialize<List<MeliAttr>>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A46D91 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd) + cateId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C041F79 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882), null, null, throwError));
	}

	public async Task<List<MeliPredictCate>> CategoryPredictor(string title)
	{
		return Deserialize<List<MeliPredictCate>>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -253477787 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4) + title, null, null, throwError: false));
	}

	public async Task<SearchClaimsResult> SearchClaims(SearchClaimsQuery query, bool throwError = false)
	{
		return Deserialize<SearchClaimsResult>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6F985201 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_da7aea68c55141aaa23181fded21b782) + query.QueryString(), null, null, throwError));
	}

	public async Task<List<ClaimInfo>> SearchOrderClaims(long orderid)
	{
		SearchClaimsQuery query = new SearchClaimsQuery
		{
			user_id = SiteUserId,
			order_id = orderid
		};
		return (await SearchClaims(query)).data;
	}

	public async Task<ClaimInfo> GetClaim(long id)
	{
		MeliApi meliApi = this;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(23, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1105F256 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315));
		defaultInterpolatedStringHandler.AppendFormatted(id);
		return Deserialize<ClaimInfo>(await meliApi.SendAsync(defaultInterpolatedStringHandler.ToStringAndClear()));
	}

	public async Task<ClaimReason> GetClaimReason(string reasonid)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x732081E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3) + reasonid;
		return Deserialize<ClaimReason>(await SendAsync(url));
	}

	public async Task<ClaimExpectedResolutions> GetClaimExpectedResolutions(long claimId)
	{
		MeliApi meliApi = this;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(44, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1374675159 ^ -233562134 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F91C0D4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a));
		return Deserialize<ClaimExpectedResolutions>(await meliApi.SendAsync(defaultInterpolatedStringHandler.ToStringAndClear()));
	}

	public async Task<ClaimExpectedResolutions> ClaimRefund(long claimId)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(51, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31345A39 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x608E2E69 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4));
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		return Deserialize<ClaimExpectedResolutions>(await SendAsync(url, HttpMethod.Post));
	}

	public async Task<PartialRefundAcailableOffers> GetClaimPartialRefundAvailableOffers(long claimId)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(55, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -445509683 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x127F552F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff));
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		return Deserialize<PartialRefundAcailableOffers>(await SendAsync(url));
	}

	public async Task<ClaimExpectedResolutions> ClaimPartialRefund(long claimId, float percentage)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(59, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E006BC5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D11FA8E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d));
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		JsonObject jsonObject = new JsonObject { [kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E5E81D1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4)] = percentage };
		return Deserialize<ClaimExpectedResolutions>(await SendAsync(url, null, jsonObject.ToString()));
	}

	public async Task<ClaimInfo> ClaimOpenDispute(long claimId)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(44, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1857DE5D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7B91271C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4d8504ccce2f465ba969392f5e065f21));
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		return Deserialize<ClaimInfo>(await SendAsync(url));
	}

	public async Task<bool> ClaimAffectsReputation(long claimId)
	{
		MeliApi meliApi = this;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(42, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236B68A8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x39996BDB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21));
		return Json.ReadString(await meliApi.SendAsync(defaultInterpolatedStringHandler.ToStringAndClear()), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C6F8540 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)).EqualIgnore(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x127F4879 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff));
	}

	public async Task<List<ClaimMessage>> GetClaimMessage(long claimId)
	{
		MeliApi meliApi = this;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(32, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67AEA6A4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE4D57F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04));
		return Deserialize<List<ClaimMessage>>(await meliApi.SendAsync(defaultInterpolatedStringHandler.ToStringAndClear()));
	}

	public async Task SendClaimMessage(long claimId, SendClaimMessage message)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(44, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3BBBEB27 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236B7C0A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8));
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		string body = message.ToJson();
		await SendAsync(url, HttpMethod.Post, body);
	}

	public async Task<string> UploadClaimAttachent(long claimId, string fileName, byte[] bytes)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(35, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x23535B1E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2DC49C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6ddd660130a6487b8fe1ae0feefd9bbe));
		string path = defaultInterpolatedStringHandler.ToStringAndClear();
		HttpRequestMessage httpRequestMessage = new HttpRequestMessage(HttpMethod.Post, BuildUrl(path));
		httpRequestMessage.Content = new MultipartFormDataContent { 
		{
			new ByteArrayContent(bytes),
			kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x273C0561 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14),
			fileName
		} };
		return Json.Get<string>(await (await SendCoreAsync(httpRequestMessage)).Content.ReadAsStringAsync(), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67AEB36A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718));
	}

	public async Task<string> DownloadClaimAttachment(long claimId, string attachid, string savePath = null)
	{
		if (savePath.IsNullOrEmpty())
		{
			savePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, attachid);
		}
		if (File.Exists(savePath))
		{
			return savePath;
		}
		try
		{
			DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(45, 2);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D11FBD0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d));
			defaultInterpolatedStringHandler.AppendFormatted(claimId);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86B58AA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb));
			defaultInterpolatedStringHandler.AppendFormatted(attachid);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86B5844 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb));
			string path = defaultInterpolatedStringHandler.ToStringAndClear();
			HttpRequestMessage request = new HttpRequestMessage(HttpMethod.Get, BuildUrl(path));
			byte[] bytes = await (await SendCoreAsync(request)).Content.ReadAsByteArrayAsync();
			string directoryName = Path.GetDirectoryName(savePath);
			if (!Directory.Exists(directoryName))
			{
				Directory.CreateDirectory(directoryName);
			}
			File.WriteAllBytes(savePath, bytes);
			return savePath;
		}
		catch (Exception)
		{
			return savePath;
		}
	}

	public async Task<string> GetClaimAttachmentUrl(long claimId, string attachid)
	{
		await CheckToken();
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(59, 3);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D11FBD0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d));
		defaultInterpolatedStringHandler.AppendFormatted(claimId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FC8C963 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913));
		defaultInterpolatedStringHandler.AppendFormatted(attachid);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x35220AB2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45));
		defaultInterpolatedStringHandler.AppendFormatted(_token);
		string path = defaultInterpolatedStringHandler.ToStringAndClear();
		return BuildUrl(path);
	}

	public async Task<MeliClipsResponse> UploadClips(MeliClipsRequest req)
	{
		string path = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x220EF407 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2) + req.cbt_item_id + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x131F8AB2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82a9ebeebdaa41eeb6924eb98e4d50c1);
		MultipartFormDataContent multipartFormDataContent = new MultipartFormDataContent();
		multipartFormDataContent.Add(new ByteArrayContent(req.content), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x73200A2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3), req.file_name);
		if (req.sites.Count > 0)
		{
			multipartFormDataContent.Add(new JsonContent(req.sites.ToJson()), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189A86F2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba));
		}
		HttpRequestMessage request = new HttpRequestMessage(HttpMethod.Post, BuildUrl(path))
		{
			Content = multipartFormDataContent
		};
		return Deserialize<MeliClipsResponse>(await (await SendCoreAsync(request)).Content.ReadAsStringAsync());
	}

	public async Task<MeliClipsInfo> GetClips(string cbtid)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4336F6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb) + cbtid + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x612BCD5F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f8d88042fa6e4eae8f42d355fd0508fd);
		return Deserialize<MeliClipsInfo>(await SendAsync(url));
	}

	public async Task<MeliPicture> GetPictureInfo(string picid)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x220EF027 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2) + picid;
		return Deserialize<MeliPicture>(await SendAsync(url));
	}

	public async Task<MeliPicture> UploadPic(byte[] bytes, string fileName, int retry = 0)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler;
		if (retry > 3)
		{
			defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(13, 1);
			defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236B7DEA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8));
			defaultInterpolatedStringHandler.AppendFormatted(retry);
			throw new Exception(defaultInterpolatedStringHandler.ToStringAndClear());
		}
		MeliPicture result = default;
		object obj;
		int num;
		try
		{
			string path = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FE9F8A1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980);
			MultipartFormDataContent multipartFormDataContent = new MultipartFormDataContent();
			string text = Path.GetExtension(new Uri(fileName).AbsolutePath);
			if (text.IsNullOrEmpty())
			{
				text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CD98C88 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479);
			}
			string fileName2 = Guid.NewGuid().ToString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C6252A0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17a4a2113da64d808d97a11eb44ba376)) + text;
			multipartFormDataContent.Add(new ByteArrayContent(bytes), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31C0DE2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339), fileName2);
			HttpRequestMessage request = new HttpRequestMessage(HttpMethod.Post, BuildUrl(path))
			{
				Content = multipartFormDataContent
			};
			result = Deserialize<MeliPicture>(await (await SendCoreAsync(request)).Content.ReadAsStringAsync());
			return result;
		}
		catch (HttpException ex)
		{
			obj = ex;
			num = 1;
		}
		if (num != 1)
		{
			return result;
		}
		HttpException ex2 = (HttpException)obj;
		TextLogger textLogger = TextLogger.Create(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F91CAFE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a));
		defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(11, 4);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B9F4E6F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c));
		defaultInterpolatedStringHandler.AppendFormatted(fileName);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3998F57D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21));
		defaultInterpolatedStringHandler.AppendFormatted(UserId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA3B3CC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22));
		defaultInterpolatedStringHandler.AppendFormatted(Aid);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x44300B2E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13));
		defaultInterpolatedStringHandler.AppendFormatted(retry);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x185AF281 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233));
		textLogger.Error(ex2, defaultInterpolatedStringHandler.ToStringAndClear());
		if (ex2.StatusCode.In(429, 500, 502))
		{
			retry++;
			await Task.Delay(600 * retry);
			return await UploadPic(bytes, fileName, retry);
		}
		ExceptionDispatchInfo.Capture((obj as Exception) ?? throw obj).Throw();
		return result;
	}

	public async Task<MeliPicture> UploadPic(string imageUrl, Action<Exception> onError = null)
	{
		_ = 1;
		try
		{
			return await UploadPic(await HttpClientEx.Create(30).GetByteArrayAsync(imageUrl), imageUrl);
		}
		catch (Exception obj)
		{
			onError?.Invoke(obj);
			return null;
		}
	}

	public async Task<string> GetDescription(string itemId, bool throwError = false)
	{
		return JsonObject.Create(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D4C663 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d) + itemId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE841943 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f), null, null, throwError)).GetString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D11E7FE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d));
	}

	public async Task<int> GetWeight(string itemId, bool throwError = false)
	{
		return Json.Get<int>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -185678843 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e) + itemId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B256094 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950), null, null, throwError), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794D0124 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6));
	}

	public async Task<MeliItemInfo> GetItemInfoFromBatch(string itemid)
	{
		MeliBatchItem[] array = await GetItems(new string[1] { itemid });
		if (array == null || array.Length == 0)
		{
			throw new Exception(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x149B82E5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04) + itemid + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8C0052 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6));
		}
		return Deserialize<MeliItemInfo>(array.First().Body.Value<string>());
	}

	public async Task<MeliItemInfo> GetItemInfo(string itemId, string include_attributes = "", bool throwError = true)
	{
		string text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x478E61EF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c) + itemId;
		if (!include_attributes.IsNullOrEmpty())
		{
			text = text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E00769D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c) + Uri.EscapeDataString(include_attributes);
		}
		return Deserialize<MeliItemInfo>(await SendAsync(text, null, null, throwError));
	}

	public async Task<MeliItemInfo> GetMarketItemInfo(string itemId, bool throwError = true)
	{
		return Deserialize<MeliItemInfo>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FE9F181 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980) + itemId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E53B495 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4), null, null, throwError));
	}

	public async Task<MeliItemInfo> GetItemIncludeAllAttrs(string itemId, bool throwError = true)
	{
		return Deserialize<MeliItemInfo>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7050694C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a21525dd11f740a2800d892a767a04e6) + itemId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EEA68AB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd), null, null, throwError));
	}

	public async Task<MeliBatchItem[]> GetItemInfoBatch(params string[] itemIds)
	{
		if (itemIds == null)
		{
			throw new ArgumentNullException(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -637693321 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb));
		}
		if (itemIds.Length == 0)
		{
			return Array.Empty<MeliBatchItem>();
		}
		if (itemIds.Length > 20)
		{
			throw new ArgumentOutOfRangeException(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -395322539 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A46F19 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd));
		}
		List<string> list = (from i in itemIds
			where !string.IsNullOrWhiteSpace(i)
			select Uri.EscapeDataString(i.Trim()) into i
			orderby i
			select i).Distinct().ToList();
		if (list.Count <= 0)
		{
			return Array.Empty<MeliBatchItem>();
		}
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3134470D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1) + string.Join(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4783BEAF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c), list);
		return JsonConvert.DeserializeObject<MeliBatchItem[]>(await SendAsync(url));
	}

	public async Task<MeliBatchItem[]> GetItems(IEnumerable<string> itemIds, bool includeAllAttrs = true)
	{
		if (itemIds == null)
		{
			throw new ArgumentNullException(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443004E2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13));
		}
		int num = itemIds.Count();
		if (num <= 0)
		{
			return Array.Empty<MeliBatchItem>();
		}
		if (num > 20)
		{
			throw new ArgumentOutOfRangeException(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F4337BA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31344727 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1));
		}
		List<string> list = (from i in itemIds
			where !string.IsNullOrWhiteSpace(i)
			select Uri.EscapeDataString(i.Trim()) into i
			orderby i
			select i).Distinct().ToList();
		if (list.Count <= 0)
		{
			return Array.Empty<MeliBatchItem>();
		}
		string text = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -783846229 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6ddd660130a6487b8fe1ae0feefd9bbe) + string.Join(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3BB6370D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa), list);
		if (includeAllAttrs)
		{
			text += kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x273C0ADB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14);
		}
		return JsonConvert.DeserializeObject<MeliBatchItem[]>(await SendAsync(text));
	}

	public async Task<MarketItemMapInfo> GetItemMapInfo(string cbtOrItemId)
	{
		return Deserialize<MarketItemMapInfo>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31C0392 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339) + cbtOrItemId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31344025 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1), null, null, throwError: false));
	}

	public async Task<InfractionInfo> GetItemInfraction(string itemId, bool throwError = true)
	{
		InfractionQuery infractionQuery = new InfractionQuery
		{
			element_id = itemId,
			element_type = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CAE7504 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22)
		};
		MeliApi meliApi = this;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(38, 2);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1594443C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120));
		defaultInterpolatedStringHandler.AppendFormatted(SiteUserId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31A4BAA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339));
		defaultInterpolatedStringHandler.AppendFormatted(infractionQuery.QueryString());
		return Deserialize<InfractionPageResult>(await meliApi.SendAsync(defaultInterpolatedStringHandler.ToStringAndClear(), null, null, throwError)).infractions?.FirstOrDefault();
	}

	public async Task<ItemPerformance> GetItemPerformance(string itemId)
	{
		return Deserialize<ItemPerformance>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE4D077 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04) + itemId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x220EF383 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2), null, null, throwError: false));
	}

	public async Task<int?> GetItemVisit(string itemId)
	{
		return JsonObject.Create(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E53B2D3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4) + itemId, null, null, throwError: false)).TryGetInt(itemId);
	}

	public async Task<SearchSiteItemsResult> SearchItemsBySeller(MeliItemQuery query = null)
	{
		query = query ?? new MeliItemQuery
		{
			limit = 50
		};
		string text = query.site_id ?? Abbr;
		if (query.seller_id.IsNullOrEmpty())
		{
			query.seller_id = SiteUserId.ToString();
		}
		query.site_id = null;
		int num = ((query.Page < 1) ? 1 : query.Page);
		query.offset = (num - 1) * query.limit;
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-536575929 ^ -2127606136 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f8d88042fa6e4eae8f42d355fd0508fd) + text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B2562BA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950) + query.QueryString();
		return Deserialize<SearchSiteItemsResult>(await SendAsync(url));
	}

	public async Task<PackMessages> GetPackMessages(long packId, MessageQuery query = null)
	{
		if (query == null)
		{
			query = new MessageQuery();
		}
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(29, 2);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x273C09E7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14));
		defaultInterpolatedStringHandler.AppendFormatted(packId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EEC2777 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd));
		defaultInterpolatedStringHandler.AppendFormatted(query.QueryString());
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		return Deserialize<PackMessages>(await SendAsync(url));
	}

	public async Task<MeliMessage> GetMessage(string id)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-713251553) ^ 0x1BB71BFE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1) + id;
		return Deserialize<PackMessages>(await SendAsync(url)).messages.FirstOrDefault();
	}

	public async Task<UnReadMessageResult> GetUnReadMessagees(UnReadMessageQuery query = null)
	{
		if (query == null)
		{
			query = new UnReadMessageQuery();
		}
		query.user_id = SiteUserId;
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1C207298 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21) + query.QueryString();
		return Deserialize<UnReadMessageResult>(await SendAsync(url));
	}

	public async Task<string> UploadMessageAttach(string fileName, string siteid, byte[] bytes)
	{
		if (siteid.IsNullOrEmpty())
		{
			siteid = Abbr;
		}
		string path = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-112040897 ^ -1426717758 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd) + siteid;
		MultipartFormDataContent multipartFormDataContent = new MultipartFormDataContent();
		multipartFormDataContent.Add(new ByteArrayContent(bytes), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D11EFEA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d), Path.GetFileName(fileName));
		HttpRequestMessage httpRequestMessage = new HttpRequestMessage(HttpMethod.Post, BuildUrl(path));
		httpRequestMessage.Content = multipartFormDataContent;
		return Json.Get<string>(await (await SendCoreAsync(httpRequestMessage)).Content.ReadAsStringAsync(), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6083C419 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4));
	}

	public async Task<string> DownloadMessageAttach(string attachid, string savePath = null)
	{
		if (savePath.IsNullOrEmpty())
		{
			savePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, attachid);
		}
		if (File.Exists(savePath))
		{
			return savePath;
		}
		try
		{
			string path = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D11E5B0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d) + attachid + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2BA63816 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e92d645a8d04335b6bb7b1de9e0a791) + Abbr;
			HttpRequestMessage request = new HttpRequestMessage(HttpMethod.Get, BuildUrl(path));
			byte[] bytes = await (await SendCoreAsync(request)).Content.ReadAsByteArrayAsync();
			string directoryName = Path.GetDirectoryName(savePath);
			if (!Directory.Exists(directoryName))
			{
				Directory.CreateDirectory(directoryName);
			}
			File.WriteAllBytes(savePath, bytes);
			return savePath;
		}
		catch (Exception)
		{
			return savePath;
		}
	}

	public async Task<string> GetMessageAttachmentUrl(string attachid)
	{
		await CheckToken();
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(71, 3);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A0BD0FF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0bd1e5ac52394aa1bf047220642aab8c));
		defaultInterpolatedStringHandler.AppendFormatted(attachid);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x233EE23C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1));
		defaultInterpolatedStringHandler.AppendFormatted(Abbr);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1898EF30 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba));
		defaultInterpolatedStringHandler.AppendFormatted(_token);
		string path = defaultInterpolatedStringHandler.ToStringAndClear();
		return BuildUrl(path);
	}

	public async Task<MeliMessage> SendMessage(SendMessage message)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(28, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43E8974B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18));
		defaultInterpolatedStringHandler.AppendFormatted(message.PackId);
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		return Deserialize<MeliMessage>(await SendAsync(url, HttpMethod.Post, message.ToJson()));
	}

	public async Task<byte[]> GetOrderInvoice(long orderId)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(28, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -2135526191 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba));
		defaultInterpolatedStringHandler.AppendFormatted(orderId);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6F985D81 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_da7aea68c55141aaa23181fded21b782));
		string path = defaultInterpolatedStringHandler.ToStringAndClear();
		HttpRequestMessage request = new HttpRequestMessage(HttpMethod.Get, BuildUrl(path));
		return await (await SendCoreAsync(request)).Content.ReadAsByteArrayAsync();
	}

	public async Task<UserInfo> GetMyInfo()
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3999733F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21);
		return Deserialize<UserInfo>(await SendAsync(url));
	}

	public async Task<ShopInfo> GetShopInfo(long shopId, bool isThrow = true)
	{
		MeliApi meliApi = this;
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(7, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E0073E7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c));
		defaultInterpolatedStringHandler.AppendFormatted(shopId);
		return Deserialize<ShopInfo>(await meliApi.SendAsync(defaultInterpolatedStringHandler.ToStringAndClear(), null, null, isThrow));
	}

	public async Task<MarketSiteInfo> GetMarketSite(long userId)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(19, 1);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CAE74EA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22));
		defaultInterpolatedStringHandler.AppendFormatted(userId);
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		return Deserialize<MarketSiteInfo>(await SendAsync(url));
	}

	public async Task<List<MarketListingQuota>> GetListingsLimit()
	{
		return Deserialize<List<MarketListingQuota>>(await SendAsync(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67AEBD16 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718)));
	}

	public async Task<UserProduct> GetUserProduct(string upId)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x478E7F67 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c) + upId;
		return Deserialize<UserProduct>(await SendVersion2Async(url));
	}

	public async Task<UserProduct> GetUserProductWithRetry(string upId, int retry = 2)
	{
		if (retry-- > 0)
		{
			try
			{
				return await GetUserProduct(upId);
			}
			catch (Exception ex)
			{
				if (ex is HttpException { StatusCode: 429 })
				{
					await Task.Delay(500);
				}
				return await GetUserProductWithRetry(upId, retry);
			}
		}
		return null;
	}

	public async Task<UserProductItemMap> GetUserProductItemMapping(string upId)
	{
		try
		{
			string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x273C0873 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14) + upId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1594481A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120);
			return Deserialize<List<UserProductItemMap>>(await SendVersion2Async(url))?.FirstOrDefault();
		}
		catch (Exception)
		{
			return null;
		}
	}

	public async Task<FamilyUserProducts> GetFamilyUserProductIds(string siteid, long familyId)
	{
		DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(31, 2);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7D9EC8F9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aa09e57b638a476287c30228f2ba6337));
		defaultInterpolatedStringHandler.AppendFormatted(siteid);
		defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67AEBA92 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718));
		defaultInterpolatedStringHandler.AppendFormatted(familyId);
		string url = defaultInterpolatedStringHandler.ToStringAndClear();
		return Deserialize<FamilyUserProducts>(await SendVersion2Async(url));
	}

	public async Task<FamilyUserProducts> GetFamilyUserProductIdsWithRetry(string siteid, long familyId, int retry = 3)
	{
		if (retry-- > 0)
		{
			try
			{
				return await GetFamilyUserProductIds(siteid, familyId);
			}
			catch (Exception ex)
			{
				if (ex is HttpException { StatusCode: 429 })
				{
					await Task.Delay(500);
				}
				return await GetFamilyUserProductIdsWithRetry(siteid, familyId, retry);
			}
		}
		return null;
	}

	public async Task<List<UserProduct>> GetUserProductsByFamilyId(string siteid, long familyId)
	{
		FamilyUserProducts familyUserProducts = await GetFamilyUserProductIds(siteid, familyId);
		List<UserProduct> list = new List<UserProduct>();
		if (familyUserProducts?.user_products_ids != null)
		{
			foreach (string user_products_id in familyUserProducts.user_products_ids)
			{
				try
				{
					UserProduct userProduct = await GetUserProductWithRetry(user_products_id);
					if (userProduct != null)
					{
						list.Add(userProduct);
					}
				}
				catch (Exception)
				{
				}
			}
		}
		return list;
	}

	public async Task<List<UserProduct>> GetAllUserProducts(UserProduct upItem)
	{
		List<UserProduct> list = new List<UserProduct> { upItem };
		if (upItem.family_id > 0)
		{
			string siteid = upItem.id.Substring(0, 3);
			FamilyUserProducts familyUserProducts = await GetFamilyUserProductIds(siteid, upItem.family_id);
			if (familyUserProducts?.user_products_ids != null)
			{
				foreach (string user_products_id in familyUserProducts.user_products_ids)
				{
					if (user_products_id == upItem.id)
					{
						continue;
					}
					try
					{
						UserProduct userProduct = await GetUserProductWithRetry(user_products_id, 5);
						if (userProduct != null)
						{
							list.Add(userProduct);
						}
					}
					catch (Exception)
					{
					}
				}
			}
		}
		return list;
	}

	public async Task<List<UserProductItemMap>> GetUPItemMappingByFamilyId(string siteid, long familyId)
	{
		List<UserProductItemMap> list = new List<UserProductItemMap>();
		FamilyUserProducts familyUserProducts = await GetFamilyUserProductIdsWithRetry(siteid, familyId);
		if (familyUserProducts == null)
		{
			return list;
		}
		if (familyUserProducts?.user_products_ids != null)
		{
			foreach (string user_products_id in familyUserProducts.user_products_ids)
			{
				UserProductItemMap userProductItemMap = await GetUserProductItemMapping(user_products_id);
				if (userProductItemMap != null)
				{
					list.Add(userProductItemMap);
				}
			}
		}
		return list;
	}

	public async Task<string> GetGlobalUserProductId(string upId)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x765A84CA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4) + upId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xE8418D5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_258c3f7a16d746e5b502228b2894957f);
		return await SendAsync(url);
	}

	public async Task<List<string>> GetLocalUserProductIds(string upId)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7C61BB02 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_db457e1da79640b7816cb6ccbfc1d724) + upId + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x612BCA47 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f8d88042fa6e4eae8f42d355fd0508fd);
		return Json.Read<List<string>>(await SendAsync(url), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43E18F97 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18));
	}

	public async Task<CreateUpResult> CreateGlobalUP(CreateUserProductModel model, bool throwError = true)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7C61B836 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_db457e1da79640b7816cb6ccbfc1d724);
		string text = await SendVersion2Async(url, HttpMethod.Post, model.ToJson(), throwError);
		if (!Json.IsJson(text))
		{
			return new CreateUpResult
			{
				Raw = text
			};
		}
		CreateUpResult createUpResult = Deserialize<CreateUpResult>(text);
		createUpResult.Raw = text;
		return createUpResult;
	}

	public async Task<CreateUpResults> CreateGlobalUPs(CreateUserProductFamilies models)
	{
		CreateUpResults list = new CreateUpResults();
		foreach (CreateUserProductModel model in models)
		{
			try
			{
				CreateUpResult createUpResult = await CreateGlobalUP(model, throwError: false);
				createUpResult.VariantKey = model.GetVariantKey();
				list.Add(createUpResult);
			}
			catch (Exception)
			{
			}
		}
		return list;
	}

	public async Task<CreateUpResult> CreateLocalUP(string upId, List<SiteToSell> sites)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2DCE0E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6ddd660130a6487b8fe1ae0feefd9bbe) + upId;
		PublishSiteToSell o = new PublishSiteToSell
		{
			sites_to_sell = sites
		};
		string text = await SendAsync(url, HttpMethod.Post, o.ToJson());
		CreateUpResult createUpResult = Deserialize<CreateUpResult>(text);
		createUpResult.Raw = text;
		return createUpResult;
	}

	public async Task UpdateUPPrice(string sitelessUpId, UpdateUserProductPrice updatePrice)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6F985B95 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_da7aea68c55141aaa23181fded21b782) + sitelessUpId;
		await SendVersion2Async(url, HttpMethod.Put, updatePrice.ToJson());
	}

	public async Task UpdateUPAttrs(string sitelessUpId, UpdateUPAttributes attrs)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794D090E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6) + sitelessUpId;
		await SendVersion2Async(url, HttpMethod.Put, attrs.ToJson());
	}

	public async Task UpdateUPWeightSize(string sitelessUpId, int weight, int[] size)
	{
		UpdateUPAttributes updateUPAttributes = new UpdateUPAttributes();
		updateUPAttributes.SetWeight(weight, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x3E131AF0 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882));
		updateUPAttributes.SetSize(size, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7B9E7258 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4d8504ccce2f465ba969392f5e065f21));
		await UpdateUPAttrs(sitelessUpId, updateUPAttributes);
	}

	public async Task UpdateUPImages(string sitelessUpId, UpdateUPImages images)
	{
		string url = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D11E554 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d) + sitelessUpId;
		await SendVersion2Async(url, HttpMethod.Put, images.ToJson());
	}

	public MeliApi(string accessToken = null, bool isCbt = true)
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
		_token = accessToken;
		_isCbt = isCbt;
	}

	public MeliApi(Record re)
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
		Aid = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x3AE0F512 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950));
		GroupId = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4788D37D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c));
		UserId = re.GetInt(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA8DE60 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22));
		ShopId = re.GetLong(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x18516DFD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233));
		SiteUserId = re.GetLong(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDBA12FE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e));
		_isCbt = re.GetBool(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C02B18D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882));
		SiteId = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -1201464039 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb));
		Abbr = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1592EFB4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120));
		AppId = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1C26DAFA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21));
		_token = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2B6190 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6ddd660130a6487b8fe1ae0feefd9bbe));
		_expire = re.GetDate(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3139B0A1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1));
		_refresh = re.Get(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7D986631 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aa09e57b638a476287c30228f2ba6337));
	}

	public async Task<string> SendAsync(string url, HttpMethod httpMethod = null, string body = null, bool throwError = true)
	{
		url = BuildUrl(url);
		if (httpMethod == null)
		{
			httpMethod = HttpMethod.Get;
		}
		HttpRequestMessage httpRequestMessage = new HttpRequestMessage(httpMethod, url);
		if (body != null)
		{
			httpRequestMessage.Content = new JsonContent(body);
		}
		return await (await SendCoreAsync(httpRequestMessage, throwError)).Content.ReadAsStringAsync();
	}

	public async Task<string> SendVersion2Async(string url, HttpMethod httpMethod = null, string body = null, bool throwError = true)
	{
		url = BuildUrl(url);
		if (httpMethod == null)
		{
			httpMethod = HttpMethod.Get;
		}
		HttpRequestMessage httpRequestMessage = new HttpRequestMessage(httpMethod, url);
		httpRequestMessage.Headers.TryAddWithoutValidation(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x15944A98 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120), kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1C1FC6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d));
		if (body != null)
		{
			httpRequestMessage.Content = new JsonContent(body);
		}
		return await (await SendCoreAsync(httpRequestMessage, throwError)).Content.ReadAsStringAsync();
	}

	private async Task<HttpResponseMessage> SendCoreAsync(HttpRequestMessage request, bool throwError = true)
	{
		request.RequestUri.ToString();
		await CheckToken();
		if (!_token.IsNullOrEmpty())
		{
			request.Headers.Authorization = new BearerToken(_token);
		}
		HttpResponseMessage rsp = await _h.SendAsync(request);
		string response = await rsp.GetString(100);
		if (!rsp.IsSuccessStatusCode & throwError)
		{
			MeliErrorParser.Instance.TryThrow((int)rsp.StatusCode, response);
		}
		return rsp;
	}

	private TData Deserialize<TData>(string json)
	{
		if (!Json.IsJson(json))
		{
			return default;
		}
		return JsonConvert.DeserializeObject<TData>(json);
	}

	private string BuildUrl(string path)
	{
		if (!path.StartsWith(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C08EA19 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882)))
		{
			return path;
		}
		return kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-781534955 ^ -590310259 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e) + path;
	}

	private async Task CheckToken()
	{
		if (Aid != 0 && !_token.IsNullOrEmpty() && !(DateTime.Now.AddMinutes(15.0) < _expire))
		{
			if (_refresh.IsNullOrEmpty())
			{
				DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(19, 1);
				defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F432AE4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb));
				defaultInterpolatedStringHandler.AppendFormatted(Aid);
				defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x127F4981 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff));
				throw new Exception(defaultInterpolatedStringHandler.ToStringAndClear());
			}
			TokenInfo tokenInfo = await RefreshToken(_refresh, _token);
			_token = tokenInfo.access_token;
			_expire = tokenInfo.ExpiresAt;
			_refresh = tokenInfo.refresh_token;
		}
	}

	public bool TokenExpired()
	{
		return DateTime.Now.AddMinutes(15.0) > _expire;
	}

	static MeliApi()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		AuthUrls = new Dictionary<string, string>
		{
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x7FD3E8E9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_17640bd4f33c49e1bebb85ad02857152)] = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x220866CB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_483ac0249f2e4059a968d91e35b6eae2),
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11092220 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315)] = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x399FE5B7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21),
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D814B9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d)] = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46F0BFB7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51),
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3BB73BBD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa)] = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D177406 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d),
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1D2A3C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d)] = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8A9528 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6),
			[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4782B3AF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c)] = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FEF6297 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980)
		};
		_h = HttpClientEx.Create(120);
	}
}
