using System;
using System.Collections.Generic;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using ZYing.Web;

namespace ZYing.Meli;

public class MeliErrorParser : ErrorMessageParser
{
	private static Dictionary<string, string> 错误对照;

	public static MeliErrorParser Instance => new MeliErrorParser();

	public Action<MeliErrorCause> ErrorHandler { get; set; }

	protected override Dictionary<string, string> Translates { get; set; }

	public override void TryThrow(int statusCode, string response)
	{
		ErrorMessage errorMessage = Parse(response);
		if (!statusCode.Between(200, 299) || errorMessage.HasError)
		{
			bool flag = CanRetry(errorMessage);
			HttpException ex = new HttpException(statusCode, errorMessage.ErrorCode, errorMessage.ErrorMsg, response);
			if (flag)
			{
				ex.SetData(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46F0BA15 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51), true);
			}
			throw ex;
		}
	}

	private bool CanRetry(ErrorMessage msg)
	{
		string value = msg.ErrorCode ?? "";
		string text = msg.ErrorMsg ?? "";
		if (text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FCE5C93 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913)))
		{
			return true;
		}
		if (value.EqualIgnore(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x765C124A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4)))
		{
			return true;
		}
		if (text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDBA2852 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e)))
		{
			return true;
		}
		if (text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1C26E32C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_84536f515c00490aa93de5c5447deb21)))
		{
			return true;
		}
		if (text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A82E06 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718)))
		{
			return true;
		}
		if (text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1773D6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d)))
		{
			return true;
		}
		if (text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3132D3F5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1)) && text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x4559F34 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c)))
		{
			return true;
		}
		if (text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C1BA889 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8853ae20538f41b5a858417f4442afe8)))
		{
			return true;
		}
		if (text.Contains(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x765C1094 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4)))
		{
			return true;
		}
		return false;
	}

	public override ErrorMessage Parse(string response)
	{
		ErrorMessage errorMessage = new ErrorMessage
		{
			RawResponse = response
		};
		if (!Json.IsJson(response))
		{
			errorMessage.ErrorMsg = response;
			return errorMessage;
		}
		MeliErrorCause meliErrorCause = Json.Read<MeliErrorCause>(response);
		ErrorHandler?.Invoke(meliErrorCause);
		errorMessage.ErrorCode = meliErrorCause.error;
		errorMessage.ErrorMsg = meliErrorCause.ToString();
		errorMessage.HasError = meliErrorCause.HasError;
		return errorMessage;
	}

	public MeliErrorParser()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		Translates = 错误对照;
		base..ctor();
	}

	static MeliErrorParser()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		错误对照 = new Dictionary<string, string>
		{
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x31A908E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1279C095 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x317ADAC4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6088A60B ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x399FE2DF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794B9128 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B23F034 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_af46a5b5cedb474285f7babb32d18950),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x2DCAEF6F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E06EC03 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -152549201 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a21525dd11f740a2800d892a767a04e6)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2B5B88 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6ddd660130a6487b8fe1ae0feefd9bbe),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x317AD4D8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x273A9581 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_35554dda7d2448649ba32ac93a833c14),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x86DC11C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c4dc1b93ffa9440986591f96c0c1f0eb)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D2419D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F45AEE2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B99DF51 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F975A1A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794B9208 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2338793E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FEF6F25 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794B9596 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794B9444 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1759067718 ^ -1438876974 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A82658 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDBA230E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8A9AC2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1F8A9B44 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_c5d3a94df2f54a1b9a979679ba9451e6)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE24BE3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B99DA71 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A200810 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C1BBFC1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8853ae20538f41b5a858417f4442afe8)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x35248B82 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5E55368D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_95a97e0027e2422199ec3b65d685d5b4)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A0D518F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0bd1e5ac52394aa1bf047220642aab8c),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1895174A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5F45B76A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_497503b5e43f458e8eab93e873a133bb),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236DF4AE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1138586725) ^ 0x78600262 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_126d3acd096442958e20a4babc5e0faa),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FCE4DB5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A2E827 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x35248E5A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11036922 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x765C01EA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_7ac61c06d562466b8704f4c1370117c4)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B99C3F5 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3132C3AD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_bbcf731df45949839e8eebe167cf5dc1)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FCE4ED9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E06FD49 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x23386A58 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ff2d6852953441efbc4475133e84f0e1),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1279DFF7 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x794B801E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1316967032) ^ 0x7079AEFF ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_001952e9043047368bc702d441752d3c)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1592C9F8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A83210 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D176FB2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -2014600060 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B99CEAD ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5AE25F4D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_11afc39c6c6e42ffbcd5c06100d5be04)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-713251553) ^ 0x1799E961 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b83e86dd962d405cad256a6b236d203a),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5C0295FB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_e0ee1e15766d4f66b64be424c2375882)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1279C1E3 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_4cac2f75f54f474aaf8e8c6d2ebd9eff),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x53628C83 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F974D4A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1592CC84 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_1435b63c8aba43d19989a88554dd1120)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-813636989) ^ 0x7BE6D462 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x2CDF0BEA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_b32cd65c86ff48fb8d106f2abf7f5479)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6C1BB351 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8853ae20538f41b5a858417f4442afe8),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-713251553) ^ 0x53C8DEA1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cf7f60f78e854a8088dc102d22b352e6)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A200500 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6F9ED6DB ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_da7aea68c55141aaa23181fded21b782)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x189518D8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7C6735A4 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_db457e1da79640b7816cb6ccbfc1d724)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A835B2 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3FEF7D05 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6e54ef65ebb146d8a11ef286af73c980)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A2E281 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A83600 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDBA3302 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D176A78 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xDBA3234 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8d66d171d78e4ac1a3beef874794bc7e),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1742346399 ^ -1690341497 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_cfb08017c5614b7bbcc362221c0a6339)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FCEB921 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6088413D ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_47c1636ee02243baa426b612de123ed4)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A21C75 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236D0754 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236D0706 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x612DBE41 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f8d88042fa6e4eae8f42d355fd0508fd)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1851B175 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_2a203f0638524f3aaf27e5c3e1951233),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x67A834D6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f696411ae93b4fd4a5c750282f8f1718)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A20FA36 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x46F05D2F ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_919dd7cda7d34057a37288631b555c51)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A21E73 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1645759935) ^ 0x7DD6F672 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1895E778 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_85b86113f83749b8965b57f4a6b46bba),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1FCEBA33 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a00098898d4a4225bca842f4eef73913)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11039FDC ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x43EEE2F1 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_00b281d5235f4b388a219ae603f8ad18)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5EEC1FC9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_0e31c6969b8e4f2aad1d571f5af793bd),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7056009C ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_a21525dd11f740a2800d892a767a04e6)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7347EF8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fbe44db3605c4574ad8406214a3619e3),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x1D1791FA ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_693bda9f31434d87b468bc88cdd69a7d)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x35247E2A ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_ba86ca7064d7426fa40034e0f348fd45),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x53A21987 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_6d4a63fb0bfc49e4bff4cdad598ceadd)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA804A8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4B993395 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_8b9ae8bfa2d246a1b5e9d10a74626d3c)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-2066591320) ^ 0x6A2E3482 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x236D00FE ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3f062cbe3aa74a10a9fdf46367ee9de8)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-(-1908959662) ^ 0x48576A75 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fa5e1389eae24359bad865a3ece14d21),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F97B642 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x317A3AA8 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_aef5dca6c7fd4f6c8d6f31da5ea0c1df),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x5A20FFC6 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_63e7983bd9c944fcba8871c47cb260ed)
			},
			{
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x11039492 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_5e7905f76bb0454a95c50ccc2bd48315),
				kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x4CA8091E ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_170b26aa155c432ab7aa98e5271bfa22)
			}
		};
	}
}
