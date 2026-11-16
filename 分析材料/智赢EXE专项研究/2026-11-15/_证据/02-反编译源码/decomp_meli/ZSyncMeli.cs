using System;
using System.Collections.Generic;
using System.Collections.Specialized;
using System.Threading.Tasks;
using NZsfdnHVTHmTqMuAXn;
using ZYing.UI;

namespace ZYing.Auto;

public static class ZSyncMeli
{
	public static Action<int, string, string> Begin;

	public static Action<string, string> Process;

	public static Action<int> Percent;

	public static Action<int, string> Done;

	public static Action<string> Output;

	public static Action<string> Error;

	public static Action Finish;

	public static Action<ZClient, string> Recieve;

	private static bool _busy;

	private static DateTime? _start;

	private static DateTime? _end;

	private static List<EItem> _list;

	public static bool IsBusy => _busy;

	public static async void Start(string startTimes, string endDates, List<EItem> list)
	{
		if (_busy)
		{
			Err(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x103DF15E ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_69972aa7c9dc4088ba7feee43919128d));
			Finish?.Invoke();
		}
		_start = (startTimes.IsNullOrEmpty() ? ((DateTime?)null) : new DateTime?(Valid.ToDate(startTimes)));
		_end = (endDates.IsNullOrEmpty() ? ((DateTime?)null) : new DateTime?(Valid.ToDate(endDates)));
		_list = list;
		_busy = true;
		for (int i = 0; i < _list.Count; i++)
		{
			if (!_busy)
			{
				break;
			}
			try
			{
				await Deal(_list[i]);
			}
			catch (Exception ex)
			{
				Err(ex.Message);
			}
		}
		_busy = false;
		Finish?.Invoke();
	}

	public static void Stop()
	{
		_busy = false;
	}

	private static async Task Deal(EItem it)
	{
		int num = Valid.ToInt(it.Key);
		WriteLine(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0xA041A25 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_505d792e9fa24a7db36aad03662e0fcd) + it.Text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x37BB91A9 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_544c7b13fb8647a7a5c14cb38647554e));
		Begin?.Invoke(num, it.Text, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(-1341199789 ^ -362061987 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_40443ba920d541b088116bf1b8a80535));
		WriteLine(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x636FB6F3 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3) + it.Text + kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x3E3B55DE ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_df600f4a0e0147a5898e6a0bb5f6c4a6));
		await 同步产品(num);
	}

	private static async Task 同步产品(int aid)
	{
		if (_busy)
		{
			ZClient zc = ZClient.Create();
			zc.Command = kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x511BE253 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_163b2de7fc8e4bb5b717eb71c068e4b7);
			NameValueCollection nv = new NameValueCollection
			{
				[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x224A6BD2 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_62aa78e010c74765857df5ce237da6e2)] = aid.ToString(),
				[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x7EACFB16 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_1c8220010bd94c9d800854148e803a6b)] = _start?.ToString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6D2DF21C ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_b3a3eb7353e14a548eba711a04d7a7e3)),
				[kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x61EE3E94 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_d0c2a484022f4cd5b4a23084a14e88ce)] = _end?.ToString(kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x6367B6B5 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_9819bf383b6e499d814d0ff5005bf7f3))
			};
			if (Recieve != null)
			{
				zc.Receiving = Recieve;
			}
			await zc.PostAsync(nv);
			if (zc.IsErr)
			{
				zc.Throw();
			}
			Done?.Invoke((int)zc.Result, kjtH3rRSqYWvB6R6Hs.EriwwTjw3(0x60F1B953 ^ <Module>{08c9c2ad-aa4e-4ae8-876d-02d123dceb6a}.m_7cfae89f1a4a471390188c613343e98f.m_a634355802034a86952a2355223f35a5));
		}
	}

	private static void WriteLine(string s)
	{
		Output?.Invoke(s);
	}

	private static void Err(string s)
	{
		Error?.Invoke(s);
	}
}
