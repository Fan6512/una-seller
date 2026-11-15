using System;
using HXK1QWpbr9xpkXuC4o;
using NZsfdnHVTHmTqMuAXn;

namespace ZYing.Meli.Model;

public class TokenInfo
{
	private int _expiresIn;

	public DateTime ExpiresAt;

	public string access_token { get; set; }

	public string refresh_token { get; set; }

	public int expires_in
	{
		get
		{
			return _expiresIn;
		}
		set
		{
			_expiresIn = value;
			ExpiresAt = DateTime.Now.AddSeconds(_expiresIn);
		}
	}

	public long user_id { get; set; }

	public string message { get; set; }

	public int status { get; set; }

	public string error { get; set; }

	public TokenInfo()
	{
		kjtH3rRSqYWvB6R6Hs.dP86a1YUY();
		OjRa8LWZubDK8CtxBF.HBece8MlL();
		base..ctor();
	}
}
