using System;
using System.Collections;
using System.Collections.Specialized;
using Hmg1hUwjXmJQ0hx3wd;
using Xmbn6iXVDuARF6WhpW;
using ZYing.Data;

namespace ZYing;

public class ZResult
{
	public static Action NoLoginHandler;

	public string Command;

	public long Result;

	public bool IsErr;

	public string Message;

	private NameValueCollection _headers;

	public RecordSet RS;

	internal static ZResult T2kGyFu1ptTKYLCJTn7d;

	public NameValueCollection Headers
	{
		get
		{
			if (_headers == null)
			{
				_headers = new NameValueCollection();
			}
			return _headers;
		}
		internal set
		{
			_headers = value;
		}
	}

	public void Throw()
	{
		if (IsErr)
		{
			throw new Exception(nvl9EQPWaAFgmhupRM9a.RyQPtm98V8J(Command, (string)YJmLc7u1R5P59pUL3Zbx(0x2426AAE6 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_0a5bf5790f514d7aaedef3a46965d602), Message, nvl9EQPWaAFgmhupRM9a.gkJPWNZA2T7));
		}
	}

	public bool Done()
	{
		if (Result == -9)
		{
			NoLoginHandler?.Invoke();
			return false;
		}
		IEnumerator enumerator = vcFHKEPf0lUDGtDIIOsa.RyQPtm98V8J(RS, vcFHKEPf0lUDGtDIIOsa.fSnPfcuQjBM);
		try
		{
			while (vvIioNPsQs26UPDdSWxV.RyQPtm98V8J(enumerator, vvIioNPsQs26UPDdSWxV.XN7PskntCpl))
			{
				MPush.Deal((Record)KWqFlDPUP88q4uU20Po0.RyQPtm98V8J(enumerator, KWqFlDPUP88q4uU20Po0.XDZPUuwY2I0));
			}
		}
		finally
		{
			if (enumerator is IDisposable disposable)
			{
				xoHMkSPsZBINt8t3oYoF.RyQPtm98V8J(disposable, xoHMkSPsZBINt8t3oYoF.oZRPsBJMmDD);
			}
		}
		return true;
	}

	public ZResult()
	{
		A8RedEce1ZSTaHmUDS.vv9ejnCAD();
		R7c2WOBSNAiVDsbUZ4.u2ulMj8ByVd();
		Message = "";
		RS = new RecordSet();
		base..ctor();
	}

	static ZResult()
	{
		A8RedEce1ZSTaHmUDS.iR3aCLEY7();
	}

	internal static bool SBOI4Ku11FWuOgVgWIkU()
	{
		return T2kGyFu1ptTKYLCJTn7d == null;
	}

	internal static ZResult i8uypZu1DAZtpuLlHowC()
	{
		return T2kGyFu1ptTKYLCJTn7d;
	}

	internal static object YJmLc7u1R5P59pUL3Zbx(int P_0)
	{
		return A8RedEce1ZSTaHmUDS.DQmUrYZ9D(P_0);
	}
}
