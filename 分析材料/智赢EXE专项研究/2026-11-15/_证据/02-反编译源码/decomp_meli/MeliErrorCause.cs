using System.Runtime.CompilerServices;
using System.Text;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;
using Newtonsoft.Json;
using ZYing.Web;

namespace ZYing.Meli;

public class MeliErrorCause
{
	[JsonProperty("status")]
	[JsonPropertyName(new string[] { "status" })]
	public int? statusCode;

	[JsonPropertyName(new string[] { "error", "code" })]
	public string error;

	[JsonPropertyName(new string[] { "message", "error_description" })]
	public string message;

	public CauseInfo cause;

	public virtual bool HasError => statusCode.Between(400, 600);

	public string TransMessage()
	{
		string transMessage = MeliErrorParser.Instance.GetTransMessage(error, message);
		if (!transMessage.IsNullOrEmpty())
		{
			return transMessage + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x443DD016 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_3862d4d1d7dc4c60a3deb6f209c23b13) + message;
		}
		return message;
	}

	public override string ToString()
	{
		if (error.IsNullOrEmpty() && message.IsNullOrEmpty())
		{
			return "";
		}
		MeliErrorParser instance = MeliErrorParser.Instance;
		StringBuilder stringBuilder = new StringBuilder(TransMessage());
		if (cause != null)
		{
			stringBuilder.AppendLine();
			int num = 1;
			foreach (string item in cause)
			{
				string text = instance.TransMessage(item);
				string value = (text.IsNullOrEmpty() ? item : (text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7F9C1622 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_82d85bc759434718a75784f94dd90a2a) + item));
				DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(4, 2);
				defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-2034326067 ^ -1052907642 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_fd98239930d443f882efd428297b997c));
				defaultInterpolatedStringHandler.AppendFormatted(num++);
				defaultInterpolatedStringHandler.AppendLiteral(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x62D9EFD9 ^ <Module>{26fe70e4-1b28-4ffe-af01-4ed86b51e71b}.m_3e5ecc1bdbb5495cab186d9c47a981b0.m_f17127e79d85492ca8e564b1e68d303d));
				defaultInterpolatedStringHandler.AppendFormatted(value);
				stringBuilder.AppendLine(defaultInterpolatedStringHandler.ToStringAndClear());
			}
		}
		return stringBuilder.ToString();
	}

	public MeliErrorCause()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}
}
