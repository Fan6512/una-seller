using System;
using System.Collections;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Collections.Specialized;
using System.Configuration;
using System.Diagnostics;
using System.IO;
using System.Net;
using System.Net.Sockets;
using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;
using System.Threading;
using System.Threading.Tasks;
using System.Windows.Forms;
using Hmg1hUwjXmJQ0hx3wd;
using Xmbn6iXVDuARF6WhpW;
using ZYing.Data;
using ZYing.Interface;
using ZYing.UI;
using ZYing.Web;

namespace ZYing;

public class ZClient
{
	public enum MType : byte
	{
		Data = 77,
		GZipHeadr = 78,
		Bytes = 87
	}

	public class IPGroup
	{
		private static List<IPGroup> _list;

		public IPAddress Address;

		public int Port;

		private ConcurrentQueue<ZClient> _pools;

		internal static IPGroup efkqWIKDtie0cjLDvD7D;

		public static IPGroup GetOrCreate(string ip)
		{
			return GetOrCreate(LIPTDTPfTJJkYYfdiLfx.RyQPtm98V8J(ip, LIPTDTPfTJJkYYfdiLfx.ICePfnfEbrA));
		}

		public static IPGroup GetOrCreate(IPAddress address)
		{
			for (int i = 0; i < _list.Count; i++)
			{
				if (_list[i].Address == address)
				{
					return _list[i];
				}
			}
			return new IPGroup(address);
		}

		public IPGroup(IPAddress ip)
		{
			A8RedEce1ZSTaHmUDS.vv9ejnCAD();
			R7c2WOBSNAiVDsbUZ4.u2ulMj8ByVd();
			Port = 7735;
			_pools = new ConcurrentQueue<ZClient>();
			base..ctor();
			Address = ip;
			_list.Add(this);
		}

		public IPGroup(IPEndPoint end)
		{
			A8RedEce1ZSTaHmUDS.vv9ejnCAD();
			R7c2WOBSNAiVDsbUZ4.u2ulMj8ByVd();
			Port = 7735;
			_pools = new ConcurrentQueue<ZClient>();
			base..ctor();
			Address = G8MTESutoU2te4KKtLSW.RyQPtm98V8J(end, G8MTESutoU2te4KKtLSW.hDKut6NHNRM);
			Port = c1wWkXutrQ87C6dk9VnL.RyQPtm98V8J(end, c1wWkXutrQ87C6dk9VnL.xYiutwgSUgd);
			_list.Add(this);
		}

		public ZClient GetPoolClient()
		{
			if (!_pools.TryDequeue(out var result))
			{
				result = CreateNew(this);
			}
			result.CleanDelegate();
			return result;
		}

		public void Enqueue(ZClient zc)
		{
			_pools.Enqueue(zc);
		}

		public override string ToString()
		{
			return nvl9EQPWaAFgmhupRM9a.RyQPtm98V8J(Dx1n9ePWIL9h3apQMSPa.RyQPtm98V8J(Address, Dx1n9ePWIL9h3apQMSPa.ykxPW2uUQor), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x45C80DAE ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5c4f0b48dde647e5b4b41ac0c211fb6c), twsXv6Pt0TqWJgLIYbQo.RyQPtm98V8J(ref Port, twsXv6Pt0TqWJgLIYbQo.VJ5PtcFXjFw), nvl9EQPWaAFgmhupRM9a.gkJPWNZA2T7);
		}

		public override int GetHashCode()
		{
			return SBeaYjP9tPqiy8uUfHWi.RyQPtm98V8J(Address, SBeaYjP9tPqiy8uUfHWi.ublP9scDJ3E) + JPIbciutM0QCPRlO35o1.RyQPtm98V8J(ref Port, JPIbciutM0QCPRlO35o1.cUCutx6jueB);
		}

		public override bool Equals(object obj)
		{
			if (!(obj is IPGroup iPGroup))
			{
				return false;
			}
			if (BRT4WKualGsvXRAy6Smf.RyQPtm98V8J(iPGroup.Address, Address, BRT4WKualGsvXRAy6Smf.ke9uaFFsCre))
			{
				return y4DLe7utTaZ52dJfuVBS.RyQPtm98V8J(ref iPGroup.Port, Port, y4DLe7utTaZ52dJfuVBS.ygSutnh2uAv);
			}
			return false;
		}

		static IPGroup()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
			A8RedEce1ZSTaHmUDS.vv9ejnCAD();
			R7c2WOBSNAiVDsbUZ4.u2ulMj8ByVd();
			_list = new List<IPGroup>();
		}

		internal static bool QANktaKDswkQXVCX44Dt()
		{
			return efkqWIKDtie0cjLDvD7D == null;
		}

		internal static IPGroup tuQG58KDWgnBfWBRDAHx()
		{
			return efkqWIKDtie0cjLDvD7D;
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <ConnectAsync>d__75 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncTaskMethodBuilder<ZResult> <>t__builder;

		public ZClient <>4__this;

		public byte[] data;

		public int trycount;

		private object <>7__wrap1;

		private int <>7__wrap2;

		private byte[] <post>5__4;

		private TaskAwaiter<int> <>u__1;

		private TaskAwaiter<byte[]> <>u__2;

		private TaskAwaiter<ZResult> <>u__3;

		private static object l2IfMjKDEpHwsb3LjWtF;

		private void MoveNext()
		{
			int num = <>1__state;
			ZClient zClient = <>4__this;
			ZResult result3;
			try
			{
				_ = 3;
				try
				{
					if ((uint)num > 2u)
					{
						if (num == 3)
						{
							goto IL_0542;
						}
						<>7__wrap2 = 0;
					}
					try
					{
						TaskAwaiter<int> awaiter = default;
						TaskAwaiter<byte[]> awaiter2;
						int num3;
						int result = default;
						byte[] result2 = default;
						switch (num)
						{
						case 1:
							awaiter = <>u__1;
							<>u__1 = default;
							num = (<>1__state = -1);
							goto IL_01d0;
						case 2:
							awaiter2 = <>u__2;
							<>u__2 = default;
							num = (<>1__state = -1);
							goto IL_02e8;
						case 0:
							awaiter = <>u__1;
							<>u__1 = default;
							num = (<>1__state = -1);
							goto IL_01ea;
						default:
							{
								zClient._busy = true;
								if (bA381mPEPUttYlfjPJ7q.RyQPtm98V8J(zClient.Command, bA381mPEPUttYlfjPJ7q.usrPEu3xg4U))
								{
									throw new Exception(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x535364B0 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_1f1df22b564a4757bdf01c98b060d27b));
								}
								if (zClient._group != null)
								{
									DateTime dateTime = A0X3HkPtQh8HhvIRv9Y9.RyQPtm98V8J(A0X3HkPtQh8HhvIRv9Y9.GhQPtkN7xja);
									TimeSpan timeSpan = sajliCP58rNuDdC0fXQB.RyQPtm98V8J(ref dateTime, zClient.ActiveTime, sajliCP58rNuDdC0fXQB.dB4P5YG8U8l);
									if (PqXddoPfXD6hAlGvjYAG.RyQPtm98V8J(ref timeSpan, PqXddoPfXD6hAlGvjYAG.CQMPfgCluIx) > 300000.0)
									{
										zClient.Close();
									}
									zClient.ActiveTime = A0X3HkPtQh8HhvIRv9Y9.RyQPtm98V8J(A0X3HkPtQh8HhvIRv9Y9.GhQPtkN7xja);
									Action<ZClient> begin = zClient.Begin;
									if (begin != null)
									{
										begin(zClient);
										goto IL_0335;
									}
									int num2 = 4;
									num3 = num2;
								}
								else
								{
									num3 = 0;
									if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_700587b288d849789a4facd655a2e021 == 0)
									{
										num3 = 1;
									}
								}
								goto IL_00c1;
							}
							IL_01d0:
							result = awaiter.GetResult();
							goto IL_016e;
							IL_016e:
							awaiter2 = zClient.Reading(result).GetAwaiter();
							if (!awaiter2.IsCompleted)
							{
								num = (<>1__state = 2);
								<>u__2 = awaiter2;
								<>t__builder.AwaitUnsafeOnCompleted(ref awaiter2, ref this);
								num3 = 0;
								if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_2242a1f0cb3a452cba50ef4a1e13ff37 == 0)
								{
									num3 = 5;
								}
								goto IL_00c1;
							}
							goto IL_02e8;
							IL_01ea:
							result = awaiter.GetResult();
							num3 = 2;
							goto IL_00c1;
							IL_02e8:
							result2 = awaiter2.GetResult();
							if (result2 == null)
							{
								if (Member.UserID == 0)
								{
									throw new Exception(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x242FA424 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_0a5bf5790f514d7aaedef3a46965d602));
								}
								zClient.Loading = false;
								result3 = zClient._ret;
								break;
							}
							goto IL_0289;
							IL_0289:
							zClient.DoParse(result2);
							zClient.ThrowNotMatch();
							zClient.DoFinish();
							<post>5__4 = null;
							goto end_IL_00a5;
							IL_0335:
							zClient.SetSerial();
							zClient.Loading = true;
							<post>5__4 = zClient.GetPost(data);
							awaiter = zClient.Init(<post>5__4).GetAwaiter();
							if (!awaiter.IsCompleted)
							{
								num = (<>1__state = 0);
								<>u__1 = awaiter;
								<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
								return;
							}
							goto IL_01ea;
							IL_00c1:
							while (true)
							{
								switch (num3)
								{
								case 5:
									return;
								case 2:
									break;
								case 3:
									goto end_IL_00c1;
								case 4:
									goto IL_0335;
								case 1:
									throw new Exception(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1B5CE00D ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_fa144e498bbe486b8f5f55f1a2c0edbf));
								default:
									goto IL_0443;
								}
								if (result == 0)
								{
									zClient.Close();
									awaiter = zClient.Init(<post>5__4, 2).GetAwaiter();
									num3 = 0;
									if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_095ec13dc2204c81987f31bb29aeab82 != 0)
									{
										num3 = 0;
									}
									continue;
								}
								goto IL_016e;
								continue;
								end_IL_00c1:
								break;
							}
							goto IL_0289;
							IL_0443:
							if (!awaiter.IsCompleted)
							{
								num = (<>1__state = 1);
								<>u__1 = awaiter;
								<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
								return;
							}
							goto IL_01d0;
						}
						goto end_IL_0066;
						end_IL_00a5:;
					}
					catch (Exception ex)
					{
						<>7__wrap1 = ex;
						<>7__wrap2 = 1;
					}
					int num4 = <>7__wrap2;
					TaskAwaiter<ZResult> awaiter3;
					if (num4 == 1)
					{
						Exception ex2 = (Exception)<>7__wrap1;
						if (bIgP8dPfrZKn5HFwjt6R.RyQPtm98V8J(ex2, bIgP8dPfrZKn5HFwjt6R.fKtPfw4Wld1) != null)
						{
							ex2 = bIgP8dPfrZKn5HFwjt6R.RyQPtm98V8J(ex2, bIgP8dPfrZKn5HFwjt6R.fKtPfw4Wld1);
						}
						if (ex2 is SocketException && trycount < 1)
						{
							awaiter3 = zClient.ConnectAsync(data, ++trycount).GetAwaiter();
							if (awaiter3.IsCompleted)
							{
								goto IL_0096;
							}
							int num5 = 0;
							if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_89cc1b5d6c3e43d589127b17bd88bcf8 != 0)
							{
								num5 = 0;
							}
							switch (num5)
							{
							case 1:
								break;
							default:
								num = (<>1__state = 3);
								<>u__3 = awaiter3;
								<>t__builder.AwaitUnsafeOnCompleted(ref awaiter3, ref this);
								return;
							}
							goto IL_0542;
						}
						zClient.DoError(ex2);
					}
					<>7__wrap1 = null;
					goto IL_0051;
					IL_0096:
					result3 = awaiter3.GetResult();
					goto end_IL_0066;
					IL_0542:
					awaiter3 = <>u__3;
					<>u__3 = default;
					num = (<>1__state = -1);
					goto IL_0096;
					end_IL_0066:;
				}
				finally
				{
					if (num < 0)
					{
						zClient._busy = false;
					}
				}
				goto end_IL_0047;
				IL_0051:
				result3 = zClient._ret;
				end_IL_0047:;
			}
			catch (Exception exception)
			{
				<>1__state = -2;
				<>t__builder.SetException(exception);
				return;
			}
			<>1__state = -2;
			<>t__builder.SetResult(result3);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			<>t__builder.SetStateMachine(stateMachine);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <ConnectAsync>d__75()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}

		internal static bool e4CQSZKDjJfEqpLQJfW7()
		{
			return l2IfMjKDEpHwsb3LjWtF == null;
		}

		internal static object AmGnm1KDbxm4KdJFrM5S()
		{
			return l2IfMjKDEpHwsb3LjWtF;
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <PostAsync>d__103 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncTaskMethodBuilder<ZResult> <>t__builder;

		public ZClient <>4__this;

		private TaskAwaiter<ZResult> <>u__1;

		private static object yINNp5KDHp9NkYExeRCc;

		private void MoveNext()
		{
			int num = <>1__state;
			ZClient zClient = <>4__this;
			ZResult result;
			try
			{
				TaskAwaiter<ZResult> awaiter;
				if (num == 0)
				{
					awaiter = <>u__1;
					<>u__1 = default;
					num = (<>1__state = -1);
				}
				else
				{
					awaiter = zClient.ConnectAsync(null).GetAwaiter();
					if (!awaiter.IsCompleted)
					{
						num = (<>1__state = 0);
						int num2 = 0;
						if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5194833652b2409baa940c831d4ad2d5 == 0)
						{
							num2 = 0;
						}
						switch (num2)
						{
						}
						<>u__1 = awaiter;
						<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
						return;
					}
				}
				result = awaiter.GetResult();
			}
			catch (Exception exception)
			{
				<>1__state = -2;
				<>t__builder.SetException(exception);
				return;
			}
			<>1__state = -2;
			<>t__builder.SetResult(result);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			<>t__builder.SetStateMachine(stateMachine);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <PostAsync>d__103()
		{
			nGcDd1KDC93BJtjEBAsa();
		}

		internal static bool MT41lQKDis4w36o4qgwZ()
		{
			return yINNp5KDHp9NkYExeRCc == null;
		}

		internal static object vtloXOKDfpmAKYmcrVnY()
		{
			return yINNp5KDHp9NkYExeRCc;
		}

		internal static void nGcDd1KDC93BJtjEBAsa()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <PostAsync>d__93 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncTaskMethodBuilder<ZResult> <>t__builder;

		public ZClient <>4__this;

		public Para p;

		private TaskAwaiter<ZResult> <>u__1;

		private static object fEd0EHKDV40hGBKwR0OV;

		private void MoveNext()
		{
			int num = <>1__state;
			ZClient zClient = <>4__this;
			ZResult result;
			try
			{
				TaskAwaiter<ZResult> awaiter;
				if (num != 0)
				{
					awaiter = zClient.PostAsync(uPBdYwPDXo36sHuZg2dw.RyQPtm98V8J(p, uPBdYwPDXo36sHuZg2dw.In4PDg5XLpP)).GetAwaiter();
					if (!awaiter.IsCompleted)
					{
						num = (<>1__state = 0);
						<>u__1 = awaiter;
						int num2 = 0;
						if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_ff6d7b6c2b7f470ab18c53a211671a61 == 0)
						{
							num2 = 0;
						}
						switch (num2)
						{
						}
						<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
						return;
					}
				}
				else
				{
					awaiter = <>u__1;
					<>u__1 = default;
					num = (<>1__state = -1);
				}
				result = awaiter.GetResult();
			}
			catch (Exception exception)
			{
				<>1__state = -2;
				<>t__builder.SetException(exception);
				return;
			}
			<>1__state = -2;
			<>t__builder.SetResult(result);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			<>t__builder.SetStateMachine(stateMachine);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <PostAsync>d__93()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}

		internal static bool wVOdi4KDdCx5mgIRUM9F()
		{
			return fEd0EHKDV40hGBKwR0OV == null;
		}

		internal static object NFRtUEKD3NqNaAE6i9AD()
		{
			return fEd0EHKDV40hGBKwR0OV;
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <PostAsync>d__95 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncTaskMethodBuilder<ZResult> <>t__builder;

		public Record[] arr;

		public ZClient <>4__this;

		private TaskAwaiter<ZResult> <>u__1;

		private static object xVibxxKD0lqjpWaZM2bE;

		private void MoveNext()
		{
			int num = <>1__state;
			ZClient zClient = <>4__this;
			ZResult result;
			try
			{
				TaskAwaiter<ZResult> awaiter;
				if (num != 0)
				{
					RecordSet recordSet = new RecordSet();
					for (int i = 0; i < arr.Length; i++)
					{
						Record record = arr[i];
						if (record != null)
						{
							nij0Skut3mCjAu3WZ9nu.RyQPtm98V8J(recordSet, record, nij0Skut3mCjAu3WZ9nu.zSbutGMYktC);
						}
					}
					awaiter = zClient.PostAsync(recordSet).GetAwaiter();
					if (awaiter.IsCompleted)
					{
						goto IL_0195;
					}
					num = (<>1__state = 0);
					<>u__1 = awaiter;
					int num2 = 1;
					if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_31cd0ff583b94a72adfb43c5b0e671c0 == 0)
					{
						num2 = 0;
					}
					switch (num2)
					{
					case 1:
						<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
						return;
					}
				}
				awaiter = <>u__1;
				<>u__1 = default;
				num = (<>1__state = -1);
				goto IL_0195;
				IL_0195:
				result = awaiter.GetResult();
			}
			catch (Exception exception)
			{
				<>1__state = -2;
				<>t__builder.SetException(exception);
				return;
			}
			<>1__state = -2;
			<>t__builder.SetResult(result);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			<>t__builder.SetStateMachine(stateMachine);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <PostAsync>d__95()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}

		internal static bool rKxErAKDcXswXjlWXG5P()
		{
			return xVibxxKD0lqjpWaZM2bE == null;
		}

		internal static object sHxL9lKDp6TAu3kB9DXF()
		{
			return xVibxxKD0lqjpWaZM2bE;
		}
	}

	public static readonly Guid UUID;

	public static readonly EItemCollection Items;

	public ZClient Next;

	public static int LastUpdate;

	public Action<int> PercentUpload;

	public bool AutoCollect;

	public string Name;

	public object Tag;

	public int Index;

	public DateTime ActiveTime;

	public string Command;

	public Action<ZClient> Begin;

	public Action<ZClient, string> Receiving;

	public Action<Exception> Failed;

	public Action<ZClient> Finish;

	private NameValueCollection _headers;

	private ZResult _ret;

	private Control _ui;

	private IPGroup _group;

	private TcpClient _tcp;

	private NetworkStream _ns;

	private byte[] _buf;

	private byte[] _tmp;

	private static int Ticks;

	private byte _headerCount;

	private int _serial;

	public int Total;

	public MType MethodType;

	private byte[] _mhead;

	private bool _busy;

	private static byte[] _key;

	public const string 泽远香港 = "8.217.74.159";

	private static string 聚石塔北京;

	public static IPGroup Server;

	public static IPGroup 老吕;

	public static IPGroup 网站;

	public static IPGroup 刊登;

	public static IPGroup 香港;

	public static IPGroup 择远;

	public static IPGroup 速卖通;

	public static IPGroup 速卖通远程;

	public static IPGroup 美客多;

	public static IPGroup N11;

	private static readonly ZClient _root;

	private static ZClient _current;

	private static Thread _thread;

	public const int TimeOutMS = 300000;

	public static bool Closing;

	internal static ZClient HQvdN4u1VHn6ornifLWZ;

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
		set
		{
			_headers = value;
		}
	}

	public long Result => _ret.Result;

	public bool IsErr => _ret.IsErr;

	public string Message => _ret.Message ?? "";

	public RecordSet RS => _ret.RS;

	public Record this[string name] => nTh7SGPU5HyNqirlBXdK.RyQPtm98V8J(_ret.RS, name, nTh7SGPU5HyNqirlBXdK.jOlPUU80rQh) ?? new Record();

	public Control UI
	{
		get
		{
			return _ui;
		}
		set
		{
			_ui = value;
			if (_ui == null)
			{
				int num = 1;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_d07f2917bfe649d6bc2f7fe0a441c80e == 0)
				{
					num = 1;
				}
				switch (num)
				{
				case 1:
					return;
				}
			}
			else if (_ui is ELayer)
			{
				if (wXGsnEPi5beMt2PRCTF9.RyQPtm98V8J(_ui, wXGsnEPi5beMt2PRCTF9.mjCPiUaj2JF) is ELayerScroll eLayerScroll)
				{
					vCSdIqPieUCN1rkJNEs1.RyQPtm98V8J(eLayerScroll, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x64DDC474 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_dd6240aeb46e481f98298eeaa3803307), vCSdIqPieUCN1rkJNEs1.glmPi9cACG7);
					_ui = eLayerScroll;
				}
				goto IL_0173;
			}
			if (_ui is EButton button)
			{
				Button = button;
			}
			goto IL_0173;
			IL_0173:
			if (Loader != null)
			{
				return;
			}
			if (!(_ui is ILoading loader))
			{
				if (wXGsnEPi5beMt2PRCTF9.RyQPtm98V8J(_ui, wXGsnEPi5beMt2PRCTF9.mjCPiUaj2JF) is ILoading loader2)
				{
					Loader = loader2;
				}
			}
			else
			{
				Loader = loader;
			}
		}
	}

	public ILoading Loader { get; set; }

	public EControl Button { get; set; }

	private bool Loading
	{
		set
		{
			if (Loader != null && !OpwJFlPiH9kynqFa274A.RyQPtm98V8J(Loader, OpwJFlPiH9kynqFa274A.GAHPiiVolSq))
			{
				VTG376PoMGKEfibgttKK.RyQPtm98V8J(Loader, value, VTG376PoMGKEfibgttKK.fgBPoxud6L2);
			}
			if (Button != null && !FbwgBhPifdAYdqsHyhGd.RyQPtm98V8J(Button, FbwgBhPifdAYdqsHyhGd.IP9PiCkmqY5))
			{
				N4jBvsPiVfIn1p969OhP.RyQPtm98V8J(Button, !value, N4jBvsPiVfIn1p969OhP.gSuPidomkfM);
			}
		}
	}

	public bool IsActive => _tcp != null;

	public int Serial => _serial;

	public bool Busy => _busy;

	public static string 工作区
	{
		get
		{
			string text = Dx1n9ePWIL9h3apQMSPa.RyQPtm98V8J(Server.Address, Dx1n9ePWIL9h3apQMSPa.ykxPW2uUQor);
			if (cYVr28PsjJv3NCy2vEkM.RyQPtm98V8J(text, C5yeLrPfeCaEPp3uUNLk.RyQPtm98V8J(Items, C5yeLrPfeCaEPp3uUNLk.uLgPf9cLSU2), cYVr28PsjJv3NCy2vEkM.spLPsbsXnhW))
			{
				return A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7EB32097 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5d2b7fa25fb44ff4b3f647983f95e6d9);
			}
			if (cYVr28PsjJv3NCy2vEkM.RyQPtm98V8J(text, lQk5hVPo0odPFusD9QVH.RyQPtm98V8J(Items[1], lQk5hVPo0odPFusD9QVH.jmlPoch7Btb), cYVr28PsjJv3NCy2vEkM.spLPsbsXnhW))
			{
				return A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7761DC92 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_2014a74f825b47e3b23645bee3edf31a);
			}
			return wZZBqoPsK6qQHL0OPRg7.RyQPtm98V8J(text, (string)KAWovlu1GlK5dRK496VD(0x58090127 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_31cd0ff583b94a72adfb43c5b0e671c0), wZZBqoPsK6qQHL0OPRg7.bZLPsLeg52J);
		}
	}

	public static ZClient Create(IPGroup remote = null)
	{
		try
		{
			if (remote == null)
			{
				remote = Server;
			}
			ZClient poolClient = remote.GetPoolClient();
			poolClient.SetGroup(remote);
			return poolClient;
		}
		catch (ConfigurationException)
		{
			ApGE7HPiTNVBbJButivp.RyQPtm98V8J(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x9F3D6D0 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f9db7669d8794e939d21bb92177af0a3), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-1978347116 ^ -912383732 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_adf85748c2d94a4ab62195edb35b4b8c), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, ApGE7HPiTNVBbJButivp.q1VPinRxQw4);
			throw new Exception(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x24BD48F8 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_6e993a12ef8e4ca986bb546bdce047f1));
		}
	}

	public void CleanDelegate()
	{
		_headers = null;
		MethodType = MType.Data;
		AutoCollect = true;
		PercentUpload = null;
		UI = null;
		int num = 0;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_2014a74f825b47e3b23645bee3edf31a == 0)
		{
			num = 0;
		}
		switch (num)
		{
		}
		Button = null;
		Loader = null;
		Command = null;
		Begin = null;
		Receiving = null;
		Failed = null;
		Finish = null;
	}

	public void Throw()
	{
		_ret.Throw();
	}

	public bool Contains(string name)
	{
		return HCC8rIPUTd0UUYRI4XC2.RyQPtm98V8J(_ret.RS, name, HCC8rIPUTd0UUYRI4XC2.NhBPUnRlBCM);
	}

	public void SetGroup(IPGroup remote)
	{
		_group = remote;
	}

	private ZClient(IPGroup remote)
	{
		A8RedEce1ZSTaHmUDS.vv9ejnCAD();
		R7c2WOBSNAiVDsbUZ4.u2ulMj8ByVd();
		AutoCollect = true;
		Name = "";
		_ret = new ZResult();
		_buf = new byte[8192];
		_tmp = new byte[8192];
		MethodType = MType.Data;
		_mhead = new byte[5];
		base..ctor();
		_group = remote;
		int num = 0;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_725165346e184a6a9e1ca06eeca2c3c1 == 0)
		{
			num = 0;
		}
		switch (num)
		{
		case 0:
			break;
		}
	}

	public void Reset()
	{
		_ret = new ZResult();
		Close();
	}

	public void Close()
	{
		try
		{
			if (_ns != null)
			{
				pkScPQPi3e1ImrhGvliS.RyQPtm98V8J(_ns, pkScPQPi3e1ImrhGvliS.OfbPiGu3Uf3);
			}
		}
		catch
		{
		}
		finally
		{
			_ns = null;
		}
		try
		{
			if (_tcp != null)
			{
				gdFZfDPiXP6APkbAnuqa.RyQPtm98V8J(_tcp, gdFZfDPiXP6APkbAnuqa.zAkPigI1PGG);
			}
		}
		catch
		{
		}
		finally
		{
			_tcp = null;
		}
	}

	private void CheckMethodType()
	{
		if (_headers != null && ESMZhyPi0f6ygNLEvCsI.RyQPtm98V8J(_headers, ESMZhyPi0f6ygNLEvCsI.d3NPiceoZ4F) > 0)
		{
			if (MethodType == MType.Data)
			{
				MethodType = MType.GZipHeadr;
			}
		}
		else if (MethodType == MType.GZipHeadr)
		{
			MethodType = MType.Data;
		}
	}

	private async Task<int> Init(byte[] post, int times = 1)
	{
		_ = 4;
		try
		{
			_mhead[0] = (byte)MethodType;
			QuyWOePflLc8SMKgL48x.RyQPtm98V8J(hJEIrPP97mXUvdTSuspT.RyQPtm98V8J(post.Length, hJEIrPP97mXUvdTSuspT.CbUP9AAlc1O), 0, _mhead, 1, 4, QuyWOePflLc8SMKgL48x.ivLPfFTSLel);
			TaskAwaiter taskAwaiter2 = default;
			TaskAwaiter taskAwaiter;
			if (_tcp == null || _ns == null)
			{
				_tcp = new TcpClient();
				taskAwaiter = FqILQkPM3oQ9xFXIdio2.RyQPtm98V8J(MaQ4bGut5AnieE0xuvY1.RyQPtm98V8J(_tcp, _group.Address, _group.Port, MaQ4bGut5AnieE0xuvY1.Y6GutUq1LI7), FqILQkPM3oQ9xFXIdio2.f12PMGUnldK);
				if (!KlwMCSuBVGSRwlo54Qc3.RyQPtm98V8J(ref taskAwaiter, KlwMCSuBVGSRwlo54Qc3.fRyuBdRK6Ca))
				{
					await taskAwaiter;
					taskAwaiter = taskAwaiter2;
					taskAwaiter2 = default;
				}
				RmENSmuB3Thq8HydVjle.RyQPtm98V8J(ref taskAwaiter, RmENSmuB3Thq8HydVjle.Lp6uBGa5HLu);
				_ns = cY7V7butevlCYUIbK3Wa.RyQPtm98V8J(_tcp, cY7V7butevlCYUIbK3Wa.I1Nut9SJ4gI);
			}
			taskAwaiter = FqILQkPM3oQ9xFXIdio2.RyQPtm98V8J(oYCnI1utHfImfIFDeija.RyQPtm98V8J(_ns, _mhead, 0, _mhead.Length, oYCnI1utHfImfIFDeija.Digutikitge), FqILQkPM3oQ9xFXIdio2.f12PMGUnldK);
			if (KlwMCSuBVGSRwlo54Qc3.RyQPtm98V8J(ref taskAwaiter, KlwMCSuBVGSRwlo54Qc3.fRyuBdRK6Ca))
			{
				goto IL_0368;
			}
			goto IL_0517;
			IL_0517:
			await taskAwaiter;
			taskAwaiter = taskAwaiter2;
			taskAwaiter2 = default;
			goto IL_0368;
			IL_0368:
			RmENSmuB3Thq8HydVjle.RyQPtm98V8J(ref taskAwaiter, RmENSmuB3Thq8HydVjle.Lp6uBGa5HLu);
			int num = 3;
			if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_8431c63ac4f045acb30300812580a306 != 0)
			{
				num = 6;
			}
			int i = default;
			int num2 = default;
			TaskAwaiter<int> taskAwaiter3 = default;
			int total = default;
			TaskAwaiter<int> taskAwaiter4 = default;
			while (true)
			{
				Action<int> percentUpload;
				switch (num)
				{
				case 2:
					taskAwaiter = FqILQkPM3oQ9xFXIdio2.RyQPtm98V8J(oYCnI1utHfImfIFDeija.RyQPtm98V8J(_ns, post, i, num2, oYCnI1utHfImfIFDeija.Digutikitge), FqILQkPM3oQ9xFXIdio2.f12PMGUnldK);
					if (!KlwMCSuBVGSRwlo54Qc3.RyQPtm98V8J(ref taskAwaiter, KlwMCSuBVGSRwlo54Qc3.fRyuBdRK6Ca))
					{
						await taskAwaiter;
						goto case 1;
					}
					goto case 5;
				default:
					return taskAwaiter3.GetResult();
				case 1:
					taskAwaiter = taskAwaiter2;
					taskAwaiter2 = default;
					num = 3;
					if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_2de3a42626df464a8848b0dea6c81a54 == 0)
					{
						num = 5;
					}
					continue;
				case 6:
					total = 0;
					i = 0;
					goto IL_01d2;
				case 5:
					RmENSmuB3Thq8HydVjle.RyQPtm98V8J(ref taskAwaiter, RmENSmuB3Thq8HydVjle.Lp6uBGa5HLu);
					i += 8192;
					goto IL_01d2;
				case 4:
					await taskAwaiter3;
					taskAwaiter3 = taskAwaiter4;
					num = 0;
					if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_29465f397d764c63b150cb36e2d80d66 == 0)
					{
						num = 0;
					}
					continue;
				case 3:
					break;
					IL_01d2:
					if (i >= post.Length)
					{
						taskAwaiter = FqILQkPM3oQ9xFXIdio2.RyQPtm98V8J(zI2DYrutfI3Sbx6E1FjL.RyQPtm98V8J(_ns, zI2DYrutfI3Sbx6E1FjL.WTlutCLLF1d), FqILQkPM3oQ9xFXIdio2.f12PMGUnldK);
						if (!KlwMCSuBVGSRwlo54Qc3.RyQPtm98V8J(ref taskAwaiter, KlwMCSuBVGSRwlo54Qc3.fRyuBdRK6Ca))
						{
							await taskAwaiter;
							taskAwaiter = taskAwaiter2;
							taskAwaiter2 = default;
						}
						RmENSmuB3Thq8HydVjle.RyQPtm98V8J(ref taskAwaiter, RmENSmuB3Thq8HydVjle.Lp6uBGa5HLu);
						taskAwaiter3 = WSAVI3utVUs0FUaBvFZG.RyQPtm98V8J(_ns, _buf, 0, _buf.Length, WSAVI3utVUs0FUaBvFZG.p2cutdqZb99).GetAwaiter();
						if (!taskAwaiter3.IsCompleted)
						{
							num = 4;
							if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_7df624928bfc47f38784b94f3192f82e == 0)
							{
								num = 0;
							}
							continue;
						}
						goto default;
					}
					num2 = zEtCCCP6MGJ0NbVFKUCS.RyQPtm98V8J(post.Length - i, 8192, zEtCCCP6MGJ0NbVFKUCS.WKYP6xqGuuL);
					total += num2;
					percentUpload = PercentUpload;
					if (percentUpload == null)
					{
						num = 0;
						if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_9f720614f9a046d49f8a1cf3c9185f69 == 0)
						{
							num = 2;
						}
						continue;
					}
					percentUpload(total * 100 / post.Length);
					goto case 2;
				}
				break;
			}
			goto IL_0517;
		}
		catch (Exception)
		{
			if (times > 1)
			{
				throw;
			}
			return 0;
		}
	}

	private unsafe async Task<ZResult> ConnectAsync(byte[] data, int trycount = 0)
	{
		_ = 3;
		try
		{
			_busy = true;
			if (bA381mPEPUttYlfjPJ7q.RyQPtm98V8J(Command, bA381mPEPUttYlfjPJ7q.usrPEu3xg4U))
			{
				throw new Exception(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x535364B0 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_1f1df22b564a4757bdf01c98b060d27b));
			}
			int num2;
			if (_group != null)
			{
				DateTime dateTime = A0X3HkPtQh8HhvIRv9Y9.RyQPtm98V8J(A0X3HkPtQh8HhvIRv9Y9.GhQPtkN7xja);
				TimeSpan timeSpan = sajliCP58rNuDdC0fXQB.RyQPtm98V8J(ref dateTime, ActiveTime, sajliCP58rNuDdC0fXQB.dB4P5YG8U8l);
				if (PqXddoPfXD6hAlGvjYAG.RyQPtm98V8J(ref timeSpan, PqXddoPfXD6hAlGvjYAG.CQMPfgCluIx) > 300000.0)
				{
					Close();
				}
				ActiveTime = A0X3HkPtQh8HhvIRv9Y9.RyQPtm98V8J(A0X3HkPtQh8HhvIRv9Y9.GhQPtkN7xja);
				Action<ZClient> begin = Begin;
				if (begin != null)
				{
					begin(this);
					goto IL_0335;
				}
				int num = 4;
				num2 = num;
			}
			else
			{
				num2 = 0;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_700587b288d849789a4facd655a2e021 == 0)
				{
					num2 = 1;
				}
			}
			goto IL_00c1;
			IL_00c1:
			int num3 = default;
			TaskAwaiter<int> taskAwaiter = default;
			byte[] post = default;
			byte[] result = default;
			TaskAwaiter<int> taskAwaiter2 = default;
			AsyncTaskMethodBuilder<ZResult> asyncTaskMethodBuilder = default;
			while (true)
			{
				TaskAwaiter<byte[]> awaiter;
				switch (num2)
				{
				case 5:
					/*Error near IL_0038: leave MoveNext - await not detected correctly*/;
				case 2:
					if (num3 == 0)
					{
						Close();
						taskAwaiter = Init(post, 2).GetAwaiter();
						num2 = 0;
						if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_095ec13dc2204c81987f31bb29aeab82 != 0)
						{
							num2 = 0;
						}
						continue;
					}
					goto IL_016e;
				case 3:
					DoParse(result);
					ThrowNotMatch();
					DoFinish();
					goto end_IL_00c1;
				case 4:
					break;
				case 1:
					throw new Exception(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1B5CE00D ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_fa144e498bbe486b8f5f55f1a2c0edbf));
				default:
					{
						if (!taskAwaiter.IsCompleted)
						{
							await taskAwaiter;
							taskAwaiter = taskAwaiter2;
							taskAwaiter2 = default;
						}
						num3 = taskAwaiter.GetResult();
						goto IL_016e;
					}
					IL_016e:
					awaiter = Reading(num3).GetAwaiter();
					if (!awaiter.IsCompleted)
					{
						asyncTaskMethodBuilder.AwaitUnsafeOnCompleted(ref awaiter, ref *(<ConnectAsync>d__75*)/*Error near IL_0264: stateMachine*/);
						num2 = 0;
						if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_2242a1f0cb3a452cba50ef4a1e13ff37 == 0)
						{
							num2 = 5;
						}
						continue;
					}
					result = awaiter.GetResult();
					if (result == null)
					{
						if (Member.UserID == 0)
						{
							throw new Exception(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x242FA424 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_0a5bf5790f514d7aaedef3a46965d602));
						}
						Loading = false;
						return _ret;
					}
					goto case 3;
				}
				goto IL_0335;
				continue;
				end_IL_00c1:
				break;
			}
			goto end_IL_0436;
			IL_0335:
			SetSerial();
			Loading = true;
			post = GetPost(data);
			num3 = await Init(post);
			num2 = 2;
			goto IL_00c1;
			end_IL_0436:;
		}
		catch (Exception ex)
		{
			if (bIgP8dPfrZKn5HFwjt6R.RyQPtm98V8J(ex, bIgP8dPfrZKn5HFwjt6R.fKtPfw4Wld1) != null)
			{
				ex = bIgP8dPfrZKn5HFwjt6R.RyQPtm98V8J(ex, bIgP8dPfrZKn5HFwjt6R.fKtPfw4Wld1);
			}
			if (ex is SocketException && trycount < 1)
			{
				ZClient zClient = this;
				int num4 = trycount + 1;
				trycount = num4;
				TaskAwaiter<ZResult> taskAwaiter3 = zClient.ConnectAsync(data, num4).GetAwaiter();
				if (!taskAwaiter3.IsCompleted)
				{
					int num5 = 0;
					if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_89cc1b5d6c3e43d589127b17bd88bcf8 != 0)
					{
						num5 = 0;
					}
					switch (num5)
					{
					case 1:
						break;
					default:
						await taskAwaiter3;
						break;
					}
					TaskAwaiter<ZResult> taskAwaiter4 = default;
					taskAwaiter3 = taskAwaiter4;
				}
				return taskAwaiter3.GetResult();
			}
			DoError(ex);
		}
		finally
		{
			_busy = false;
		}
		return _ret;
	}

	private async Task<byte[]> Reading(int readed)
	{
		_headerCount = 0;
		int len = 0;
		int offset = 0;
		short slen = 0;
		MemoryStream ms = null;
		try
		{
			TaskAwaiter<int> taskAwaiter = default;
			TaskAwaiter<int> taskAwaiter2 = default;
			int num2 = default;
			int num3 = default;
			while (true)
			{
				IL_027d:
				if (readed <= 0)
				{
					goto IL_039a;
				}
				int num;
				if (Receiving != null)
				{
					if (offset > 0)
					{
						num = 5;
						goto IL_00b3;
					}
					goto IL_0449;
				}
				goto IL_0640;
				IL_0198:
				if (_ns == null)
				{
					num = 6;
				}
				else
				{
					taskAwaiter = WSAVI3utVUs0FUaBvFZG.RyQPtm98V8J(_ns, _buf, 0, _buf.Length, WSAVI3utVUs0FUaBvFZG.p2cutdqZb99).GetAwaiter();
					if (taskAwaiter.IsCompleted)
					{
						goto IL_05ed;
					}
					await taskAwaiter;
					taskAwaiter = taskAwaiter2;
					taskAwaiter2 = default;
					num = 1;
					if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_28721135bed84ee2ab8eccd2f8df395b != 0)
					{
						num = 2;
					}
				}
				goto IL_00b3;
				IL_00b3:
				while (true)
				{
					switch (num)
					{
					case 1:
						break;
					case 3:
						goto end_IL_00b3;
					default:
						goto IL_0456;
					case 4:
						readed = num2;
						goto IL_027d;
					case 6:
						num2 = 0;
						goto case 4;
					case 5:
						goto IL_053b;
					case 2:
						goto IL_05ed;
					}
					if (_buf[offset] != 119 || _buf[offset + 1] != 53)
					{
						goto IL_0640;
					}
					slen = ooZeoout0HKLxZJnMJKd.RyQPtm98V8J(_buf, offset + 2, ooZeoout0HKLxZJnMJKd.TK1utc4rkTe);
					offset += 4;
					num3 = readed - 4;
					if (slen > num3)
					{
						readed -= 4;
						QuyWOePflLc8SMKgL48x.RyQPtm98V8J(_buf, offset, _tmp, 0, num3, QuyWOePflLc8SMKgL48x.ivLPfFTSLel);
						num = 3;
						continue;
					}
					goto IL_0341;
					continue;
					end_IL_00b3:
					break;
				}
				offset = slen - num3;
				goto IL_0198;
				IL_053b:
				QuyWOePflLc8SMKgL48x.RyQPtm98V8J(_buf, 0, _tmp, slen - offset, offset, QuyWOePflLc8SMKgL48x.ivLPfFTSLel);
				string s = oBsiiRPfZB6dPZNCD0XO.RyQPtm98V8J(fgIXDhPsDG7ytC1qXiFj.RyQPtm98V8J(fgIXDhPsDG7ytC1qXiFj.rKbPsReQFvE), _tmp, 0, slen, oBsiiRPfZB6dPZNCD0XO.ISaPfBgv2EK);
				DoReceiving(s);
				readed -= offset;
				goto IL_0449;
				IL_0449:
				int num4 = offset;
				goto IL_04f5;
				IL_04f5:
				if (readed > num4 + 4)
				{
					num = 1;
					if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5c4f0b48dde647e5b4b41ac0c211fb6c == 0)
					{
						num = 1;
					}
					goto IL_00b3;
				}
				goto IL_0640;
				IL_0456:
				offset = 0;
				goto IL_036f;
				IL_036f:
				if (Pe8srKPdWAfYUJGOXGlN.RyQPtm98V8J(ms, Pe8srKPdWAfYUJGOXGlN.BNwPdExZEPb) < len)
				{
					goto IL_0198;
				}
				goto IL_039a;
				IL_0640:
				if (ms == null)
				{
					int num5 = offset;
					if (_buf[num5++] == 99)
					{
						_headerCount = _buf[num5++];
						len = lIFL86PHWkl1RPS6gTfM.RyQPtm98V8J(_buf, num5, lIFL86PHWkl1RPS6gTfM.MgrPHE9lK2n);
						num5 += 4;
						ms = new MemoryStream();
						oOBHlrP98gwmHg7AMQ16.RyQPtm98V8J(ms, _buf, num5, readed - (num5 - offset), oOBHlrP98gwmHg7AMQ16.m1AP9YNEnxi);
						num = 0;
						if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_095ec13dc2204c81987f31bb29aeab82 != 0)
						{
							num = 0;
						}
						goto IL_00b3;
					}
					break;
				}
				oOBHlrP98gwmHg7AMQ16.RyQPtm98V8J(ms, _buf, offset, readed, oOBHlrP98gwmHg7AMQ16.m1AP9YNEnxi);
				goto IL_036f;
				IL_039a:
				MemoryStream memoryStream = ms;
				return (memoryStream != null) ? so7aOIPrMECVdyYotTJu.RyQPtm98V8J(memoryStream, so7aOIPrMECVdyYotTJu.bbkPrxlqT68) : null;
				IL_05ed:
				num2 = taskAwaiter.GetResult();
				num = 4;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_d0b0e564dc3c4d449b8fabfc55da559f != 0)
				{
					num = 4;
				}
				goto IL_00b3;
				IL_0341:
				string s2 = oBsiiRPfZB6dPZNCD0XO.RyQPtm98V8J(fgIXDhPsDG7ytC1qXiFj.RyQPtm98V8J(fgIXDhPsDG7ytC1qXiFj.rKbPsReQFvE), _buf, offset, slen, oBsiiRPfZB6dPZNCD0XO.ISaPfBgv2EK);
				DoReceiving(s2);
				readed -= slen + 4;
				if (readed > 0)
				{
					if (readed >= 4)
					{
						num4 = 0;
						offset += slen;
						goto IL_04f5;
					}
					QuyWOePflLc8SMKgL48x.RyQPtm98V8J(_buf, offset, _tmp, 0, readed, QuyWOePflLc8SMKgL48x.ivLPfFTSLel);
					offset = slen - readed;
				}
				else
				{
					offset = 0;
				}
				goto IL_0198;
			}
			throw new Exception(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6A28BD8F ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_28721135bed84ee2ab8eccd2f8df395b));
		}
		finally
		{
			MemoryStream memoryStream2 = ms;
			if (memoryStream2 != null)
			{
				pkScPQPi3e1ImrhGvliS.RyQPtm98V8J(memoryStream2, pkScPQPi3e1ImrhGvliS.OfbPiGu3Uf3);
			}
		}
	}

	private void SetSerial()
	{
		_serial = d8U9k1PipG2m6XG0qg3w.RyQPtm98V8J(ref Ticks, d8U9k1PipG2m6XG0qg3w.d6jPi1w9HEU);
		if (_ui != null && _ui is EControl eControl)
		{
			DTglwkPiD4k4uacqP2yG.RyQPtm98V8J(eControl, _serial, DTglwkPiD4k4uacqP2yG.NSVPiR3m7It);
		}
	}

	private void ThrowNotMatch()
	{
		if (Closing)
		{
			throw new IgnoreException(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7C44FB7A ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_36369014ecd545c39ede59c60a92f687));
		}
		if (_ui != null && _ui is EControl eControl && nFu5XTPi76rZYVAwNsyJ.RyQPtm98V8J(eControl, nFu5XTPi76rZYVAwNsyJ.EdUPiAEZo9V) != _serial)
		{
			string[] array = new string[7]
			{
				A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x62AFC01A ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_fa27fd6b6f2a44f6a9954a4ca58fab8d),
				twsXv6Pt0TqWJgLIYbQo.RyQPtm98V8J(ref _serial, twsXv6Pt0TqWJgLIYbQo.VJ5PtcFXjFw),
				A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x79FEB4D4 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_304a713a41a54d23b34fd4cec90492ed),
				fT6dnUPi8Vr0BSNVc98k.RyQPtm98V8J(eControl, fT6dnUPi8Vr0BSNVc98k.x3HPiY2kgdt),
				A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7EB327C9 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5d2b7fa25fb44ff4b3f647983f95e6d9),
				null,
				null
			};
			int num = nFu5XTPi76rZYVAwNsyJ.RyQPtm98V8J(eControl, nFu5XTPi76rZYVAwNsyJ.EdUPiAEZo9V);
			array[5] = twsXv6Pt0TqWJgLIYbQo.RyQPtm98V8J(ref num, twsXv6Pt0TqWJgLIYbQo.VJ5PtcFXjFw);
			array[6] = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7B58A4AB ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_27dc7b16621540c3867a9b2973100be1);
			throw new IgnoreException(cL88WfPtDRLFVoQ7PO54.RyQPtm98V8J(array, cL88WfPtDRLFVoQ7PO54.YLdPtRJSvas));
		}
	}

	private void DoReceiving(string s)
	{
		if (Receiving == null || (_ui != null && (FbwgBhPifdAYdqsHyhGd.RyQPtm98V8J(_ui, FbwgBhPifdAYdqsHyhGd.IP9PiCkmqY5) || Ff3ISkPiSdHR3MksQMnm.RyQPtm98V8J(_ui, Ff3ISkPiSdHR3MksQMnm.FrUPiJGwBLS))))
		{
			return;
		}
		try
		{
			if (_ui != null && h7E2S9PizyPKkYO9aNhK.RyQPtm98V8J(_ui, h7E2S9PizyPKkYO9aNhK.lQUPfOxEb9e))
			{
				_ui.Invoke(new Action<ZClient, string>(Receiving.Invoke), this, s);
			}
			else
			{
				Receiving(this, s);
			}
		}
		catch
		{
		}
	}

	private byte[] GetPost(byte[] data)
	{
		CheckMethodType();
		MemoryStream memoryStream = new MemoryStream();
		try
		{
			gLNAnPPfIOtyL5fxa1gr.RyQPtm98V8J(memoryStream, LastUpdate, gLNAnPPfIOtyL5fxa1gr.wQaPf2Cymj8);
			N0t65kPfyLfqtMqiwD6H.RyQPtm98V8J(memoryStream, UUID, N0t65kPfyLfqtMqiwD6H.HPnPfhBYH0u);
			N0t65kPfyLfqtMqiwD6H.RyQPtm98V8J(memoryStream, Member.Token, N0t65kPfyLfqtMqiwD6H.HPnPfhBYH0u);
			iaw0NnPfaYR4HeAoew68.RyQPtm98V8J(memoryStream, Command, iaw0NnPfaYR4HeAoew68.MRCPfNex2bF);
			if (MethodType == MType.GZipHeadr)
			{
				byte[] array = NameValuesToBytes(_headers);
				int num = 0;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_107ad496e3da4f6ca24971391e57f654 != 0)
				{
					num = 0;
				}
				switch (num)
				{
				}
				oOBHlrP98gwmHg7AMQ16.RyQPtm98V8J(memoryStream, hJEIrPP97mXUvdTSuspT.RyQPtm98V8J(array.Length, hJEIrPP97mXUvdTSuspT.CbUP9AAlc1O), 0, 4, oOBHlrP98gwmHg7AMQ16.m1AP9YNEnxi);
				oOBHlrP98gwmHg7AMQ16.RyQPtm98V8J(memoryStream, array, 0, array.Length, oOBHlrP98gwmHg7AMQ16.m1AP9YNEnxi);
			}
			if (data != null)
			{
				oOBHlrP98gwmHg7AMQ16.RyQPtm98V8J(memoryStream, data, 0, data.Length, oOBHlrP98gwmHg7AMQ16.m1AP9YNEnxi);
			}
			return so7aOIPrMECVdyYotTJu.RyQPtm98V8J(memoryStream, so7aOIPrMECVdyYotTJu.bbkPrxlqT68);
		}
		finally
		{
			if (memoryStream != null)
			{
				xoHMkSPsZBINt8t3oYoF.RyQPtm98V8J(memoryStream, xoHMkSPsZBINt8t3oYoF.oZRPsBJMmDD);
			}
		}
	}

	private static byte[] NameValuesToBytes(NameValueCollection nv)
	{
		MemoryStream memoryStream = new MemoryStream();
		try
		{
			IEnumerator enumerator = NprKEuPfPKQ3hHUgW11i.RyQPtm98V8J(nv, NprKEuPfPKQ3hHUgW11i.rAxPfumV0c7);
			try
			{
				while (vvIioNPsQs26UPDdSWxV.RyQPtm98V8J(enumerator, vvIioNPsQs26UPDdSWxV.XN7PskntCpl))
				{
					string text = (string)KWqFlDPUP88q4uU20Po0.RyQPtm98V8J(enumerator, KWqFlDPUP88q4uU20Po0.XDZPUuwY2I0);
					string text2 = lIGa0DP6feGC5vWvyuZ1.RyQPtm98V8J(nv, text, lIGa0DP6feGC5vWvyuZ1.U7CP6Cwrcgr);
					if (text2 != null)
					{
						iaw0NnPfaYR4HeAoew68.RyQPtm98V8J(memoryStream, text, iaw0NnPfaYR4HeAoew68.MRCPfNex2bF);
						gEaQufPfKbIaQ9CRTWc9.RyQPtm98V8J(memoryStream, text2, gEaQufPfKbIaQ9CRTWc9.VXpPfL5Ox3J);
					}
				}
			}
			finally
			{
				if (enumerator is IDisposable disposable)
				{
					xoHMkSPsZBINt8t3oYoF.RyQPtm98V8J(disposable, xoHMkSPsZBINt8t3oYoF.oZRPsBJMmDD);
				}
			}
			return so7aOIPrMECVdyYotTJu.RyQPtm98V8J(memoryStream, so7aOIPrMECVdyYotTJu.bbkPrxlqT68);
		}
		finally
		{
			if (memoryStream != null)
			{
				xoHMkSPsZBINt8t3oYoF.RyQPtm98V8J(memoryStream, xoHMkSPsZBINt8t3oYoF.oZRPsBJMmDD);
			}
		}
	}

	private static byte[] Encrypt(byte[] arr)
	{
		for (int i = 0; i < arr.Length; i++)
		{
			arr[i] ^= _key[i % _key.Length];
		}
		return arr;
	}

	private static byte[] Decrypt(byte[] arr)
	{
		for (int i = 0; i < arr.Length; i++)
		{
			arr[i] ^= _key[i % _key.Length];
		}
		return arr;
	}

	private void DoParse(byte[] back)
	{
		try
		{
			long result = default;
			bool isErr = default;
			int num = default;
			string message = default;
			RecordSet recordSet = default;
			int num3;
			if (back.Length >= 9)
			{
				num = 0;
				LastUpdate = lIFL86PHWkl1RPS6gTfM.RyQPtm98V8J(back, num, lIFL86PHWkl1RPS6gTfM.MgrPHE9lK2n);
				num += 4;
				byte[] array = new byte[16];
				QuyWOePflLc8SMKgL48x.RyQPtm98V8J(back, num, array, 0, 16, QuyWOePflLc8SMKgL48x.ivLPfFTSLel);
				Guid guid = new Guid(array);
				num += 16;
				QuyWOePflLc8SMKgL48x.RyQPtm98V8J(back, num, array, 0, 16, QuyWOePflLc8SMKgL48x.ivLPfFTSLel);
				Guid guid2 = new Guid(array);
				num += 16;
				if (roRQQUPbQCCELedUHHdc.RyQPtm98V8J(guid, UUID, roRQQUPbQCCELedUHHdc.hsCPbkNZJ87) || roRQQUPbQCCELedUHHdc.RyQPtm98V8J(guid2, Member.Token, roRQQUPbQCCELedUHHdc.hsCPbkNZJ87))
				{
					throw new NoLoginEx(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x9F3D736 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f9db7669d8794e939d21bb92177af0a3));
				}
				result = dUvhN3Pf4ehuBvUDLBiG.RyQPtm98V8J(back, num, dUvhN3Pf4ehuBvUDLBiG.I69PfmC619S);
				num += 8;
				isErr = yulh1iPfQYM6PNnUhxQt.RyQPtm98V8J(back[num++], yulh1iPfQYM6PNnUhxQt.LZcPfkFZEsG);
				int num2 = lIFL86PHWkl1RPS6gTfM.RyQPtm98V8J(back, num, lIFL86PHWkl1RPS6gTfM.MgrPHE9lK2n);
				num += 4;
				if (num2 > back.Length - num || num2 < 0)
				{
					throw new IgnoreException();
				}
				if (num2 > 0)
				{
					string text = oBsiiRPfZB6dPZNCD0XO.RyQPtm98V8J(fgIXDhPsDG7ytC1qXiFj.RyQPtm98V8J(fgIXDhPsDG7ytC1qXiFj.rKbPsReQFvE), back, num, num2, oBsiiRPfZB6dPZNCD0XO.ISaPfBgv2EK);
					num += num2;
					message = text;
				}
				else
				{
					message = "";
				}
				_headers = null;
				if (_headerCount > 0)
				{
					_headers = new NameValueCollection();
					WriteNameValues(_headers, back, _headerCount, ref num);
				}
				recordSet = new RecordSet();
				num3 = 0;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_69abf8cc2cd9444db21ebabb4b202604 != 0)
				{
					num3 = 0;
				}
			}
			else
			{
				num3 = 2;
			}
			while (true)
			{
				switch (num3)
				{
				case 1:
					_ret = new ZResult
					{
						Command = Command,
						Result = result,
						IsErr = isErr,
						Message = message,
						Headers = _headers,
						RS = recordSet
					};
					return;
				case 2:
					throw new IgnoreException();
				default:
					ORnnjFPfqLWAEOemWljg.RyQPtm98V8J(recordSet, back, num, ORnnjFPfqLWAEOemWljg.SQxPfvwaXjI);
					num3 = 1;
					if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_023613f232e14273961a9409b493fa36 != 0)
					{
						num3 = 1;
					}
					break;
				}
			}
		}
		catch (Exception)
		{
			Reset();
			throw new IgnoreException();
		}
	}

	internal static void WriteNameValues(NameValueCollection nv, byte[] data, int headerCount, ref int offset)
	{
		if (data == null || data.Length == 0)
		{
			return;
		}
		MemoryStream memoryStream = new MemoryStream(data);
		try
		{
			if (offset > 0)
			{
				int num = 0;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_304a713a41a54d23b34fd4cec90492ed == 0)
				{
					num = 0;
				}
				switch (num)
				{
				}
				LmBNcmP6lMu862UZJ7vI.RyQPtm98V8J(memoryStream, offset, SeekOrigin.Begin, LmBNcmP6lMu862UZJ7vI.d25P6FHFmaF);
			}
			while (headerCount > 0)
			{
				string text = MhPVkcPftKL1LfTPJg8Y.RyQPtm98V8J(memoryStream, MhPVkcPftKL1LfTPJg8Y.gYZPfsIGC7w);
				if (i97HtxPjymdGiLMljiyR.RyQPtm98V8J(text, i97HtxPjymdGiLMljiyR.CrZPjhYeUf2) == 0)
				{
					break;
				}
				bRCg2LPfjco89dAY95bx.RyQPtm98V8J(nv, text, PJNUnQPfWJDbACr2xvAt.RyQPtm98V8J(memoryStream, PJNUnQPfWJDbACr2xvAt.mqlPfEmPvh4), bRCg2LPfjco89dAY95bx.Fi4Pfbhtu5l);
				headerCount--;
			}
			offset = (int)ImWVEBPfoGEZskwS7fc7.RyQPtm98V8J(memoryStream, ImWVEBPfoGEZskwS7fc7.qsAPf6mlCFL);
		}
		finally
		{
			if (memoryStream != null)
			{
				xoHMkSPsZBINt8t3oYoF.RyQPtm98V8J(memoryStream, xoHMkSPsZBINt8t3oYoF.oZRPsBJMmDD);
			}
		}
	}

	private void DoError(Exception ex)
	{
		Loading = false;
		Reset();
		int num = 0;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_573ffbac905f4c0c9da5656517682a80 == 0)
		{
			num = 0;
		}
		switch (num)
		{
		}
		if (AutoCollect)
		{
			GCollect();
		}
		if (!(ex is IgnoreException ex2))
		{
			ShowErr(ex);
		}
		else if (Finish == null)
		{
			throw ex2;
		}
	}

	private void ShowErr(Exception ex)
	{
		try
		{
			if (bIgP8dPfrZKn5HFwjt6R.RyQPtm98V8J(ex, bIgP8dPfrZKn5HFwjt6R.fKtPfw4Wld1) != null)
			{
				ex = bIgP8dPfrZKn5HFwjt6R.RyQPtm98V8J(ex, bIgP8dPfrZKn5HFwjt6R.fKtPfw4Wld1);
			}
			if (ex is SocketException ex2)
			{
				int num = 0;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_4807bd5c4dc84fc195e0105a5aa96c33 != 0)
				{
					num = 0;
				}
				switch (num)
				{
				}
				Close();
				if (dFKRabPfM8ZKuDh0tFCw.RyQPtm98V8J(ex2, dFKRabPfM8ZKuDh0tFCw.vAIPfx5929i) == 10057)
				{
					return;
				}
				int num2 = dFKRabPfM8ZKuDh0tFCw.RyQPtm98V8J(ex2, dFKRabPfM8ZKuDh0tFCw.vAIPfx5929i);
				ex = new Exception(wZZBqoPsK6qQHL0OPRg7.RyQPtm98V8J(twsXv6Pt0TqWJgLIYbQo.RyQPtm98V8J(ref num2, twsXv6Pt0TqWJgLIYbQo.VJ5PtcFXjFw), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6D3A29FE ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_7cb36848e8df46f6a04c61a49eb22da2), wZZBqoPsK6qQHL0OPRg7.bZLPsLeg52J), (Exception?)(object)ex2);
			}
			if (Failed != null)
			{
				Failed(ex);
				return;
			}
			_ret.IsErr = true;
			_ret.Message = WnnODyPj5l1JgCPKaH9d.RyQPtm98V8J(ex, WnnODyPj5l1JgCPKaH9d.OW0PjUZ1fMW);
			OnFinish();
		}
		catch (Exception)
		{
		}
	}

	private void DoHeader()
	{
		if (_headers != null)
		{
			MPush.GetPushData(lIGa0DP6feGC5vWvyuZ1.RyQPtm98V8J(_headers, (string)KAWovlu1GlK5dRK496VD(0xD0E2ED9 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_c006ecad44af4f91bbb59ddc7246133d), lIGa0DP6feGC5vWvyuZ1.U7CP6Cwrcgr));
		}
	}

	private void DoFinish()
	{
		Loading = false;
		try
		{
			if (!nigPZEu1X10BjV4U25bO(_ret) || !_ret.IsErr)
			{
				DoHeader();
				OnFinish();
				return;
			}
			throw new Exception(wZZBqoPsK6qQHL0OPRg7.RyQPtm98V8J(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x24BD49E2 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_6e993a12ef8e4ca986bb546bdce047f1), _ret.Message, wZZBqoPsK6qQHL0OPRg7.bZLPsLeg52J));
		}
		catch (Exception ex)
		{
			ShowErr(ex);
		}
		finally
		{
			if (AutoCollect)
			{
				GCollect();
			}
		}
	}

	private void OnFinish()
	{
		if (Finish != null)
		{
			if (_ui != null && h7E2S9PizyPKkYO9aNhK.RyQPtm98V8J(_ui, h7E2S9PizyPKkYO9aNhK.lQUPfOxEb9e))
			{
				_ui.Invoke(new Action<ZClient>(Finish.Invoke), this);
			}
			else
			{
				Finish(this);
			}
		}
	}

	public void GCollect()
	{
		Yk9IQiu1gp7vJ0nH0CMS(_group, this);
	}

	[AsyncStateMachine(typeof(<PostAsync>d__93))]
	public Task<ZResult> PostAsync(Para p)
	{
		// ILSpy could not decompile this member: Error decompiling @060014D7 ZYing.ZClient.PostAsync: Index was out of range. Must be non-negative and less than the size of the collection. (Parameter 'index')
		// Please report the exception below,
		// along with the assembly it came from, at https://github.com/icsharpcode/ILSpy/issues/new
		// Error decompiling @060014D7 ZYing.ZClient.PostAsync
		// in assembly "C:/Users/ASUS/AppData/Local/Temp/zying-rea-20261115/named\ZYing.Work.dll"
		//  ---> System.ArgumentOutOfRangeException: Index was out of range. Must be non-negative and less than the size of the collection. (Parameter 'index')
		//    at System.ThrowHelper.ThrowArgumentOutOfRange_IndexMustBeLessException() + 0x1b
		//    at ICSharpCode.Decompiler.IL.ControlFlow.AsyncAwaitDecompiler.AnalyzeAwaitBlock(Block, ILVariable&, IField&, Int32&, Int32&) + 0x593
		//    at ICSharpCode.Decompiler.IL.ControlFlow.AsyncAwaitDecompiler.AnalyzeStateMachine(ILFunction) + 0x248
		//    at ICSharpCode.Decompiler.IL.ControlFlow.AsyncAwaitDecompiler.Run(ILFunction, ILTransformContext) + 0xdb
		//    at ICSharpCode.Decompiler.IL.ILFunction.RunTransforms(IEnumerable`1, ILTransformContext) + 0x75
		//    at ICSharpCode.Decompiler.CSharp.CSharpDecompiler.DecompileBody(IMethod, EntityDeclaration, DecompileRun, ITypeResolveContext, ExtensionInfo) + 0x3b9
		// -- continuing with outer exception (ICSharpCode.Decompiler.DecompilerException) --
		// 
	}

	public async Task<ZResult> PostAsync(NameValueCollection nv)
	{
		byte[] data = NameValuesToBytes(nv);
		int num = 0;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_9f720614f9a046d49f8a1cf3c9185f69 != 0)
		{
			num = 0;
		}
		return num switch
		{
			_ => await PostAsync(data), 
		};
	}

	[AsyncStateMachine(typeof(<PostAsync>d__95))]
	public Task<ZResult> PostAsync(params Record[] arr)
	{
		// ILSpy could not decompile this member: Error decompiling @060014D9 ZYing.ZClient.PostAsync: Index was out of range. Must be non-negative and less than the size of the collection. (Parameter 'index')
		// Please report the exception below,
		// along with the assembly it came from, at https://github.com/icsharpcode/ILSpy/issues/new
		// Error decompiling @060014D9 ZYing.ZClient.PostAsync
		// in assembly "C:/Users/ASUS/AppData/Local/Temp/zying-rea-20261115/named\ZYing.Work.dll"
		//  ---> System.ArgumentOutOfRangeException: Index was out of range. Must be non-negative and less than the size of the collection. (Parameter 'index')
		//    at System.ThrowHelper.ThrowArgumentOutOfRange_IndexMustBeLessException() + 0x1b
		//    at ICSharpCode.Decompiler.IL.ControlFlow.AsyncAwaitDecompiler.AnalyzeAwaitBlock(Block, ILVariable&, IField&, Int32&, Int32&) + 0x593
		//    at ICSharpCode.Decompiler.IL.ControlFlow.AsyncAwaitDecompiler.AnalyzeStateMachine(ILFunction) + 0x248
		//    at ICSharpCode.Decompiler.IL.ControlFlow.AsyncAwaitDecompiler.Run(ILFunction, ILTransformContext) + 0xdb
		//    at ICSharpCode.Decompiler.IL.ILFunction.RunTransforms(IEnumerable`1, ILTransformContext) + 0x75
		//    at ICSharpCode.Decompiler.CSharp.CSharpDecompiler.DecompileBody(IMethod, EntityDeclaration, DecompileRun, ITypeResolveContext, ExtensionInfo) + 0x3b9
		// -- continuing with outer exception (ICSharpCode.Decompiler.DecompilerException) --
		// 
	}

	public async Task<ZResult> PostAsync(RecordSet rs)
	{
		TaskAwaiter<ZResult> taskAwaiter = PostAsync(WqgMUautXptYKu7bOnf3.RyQPtm98V8J(rs, WqgMUautXptYKu7bOnf3.S7sutg6EtBQ)).GetAwaiter();
		if (!taskAwaiter.IsCompleted)
		{
			await taskAwaiter;
			TaskAwaiter<ZResult> taskAwaiter2 = default;
			taskAwaiter = taskAwaiter2;
			int num = 0;
			if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_9f720614f9a046d49f8a1cf3c9185f69 == 0)
			{
				num = 0;
			}
			switch (num)
			{
			}
		}
		return taskAwaiter.GetResult();
	}

	public async Task<ZResult> PostAsync(string name, bool value)
	{
		TaskAwaiter<ZResult> taskAwaiter = PostAsync(name, value ? 1 : 0).GetAwaiter();
		if (!taskAwaiter.IsCompleted)
		{
			await taskAwaiter;
			TaskAwaiter<ZResult> taskAwaiter2 = default;
			taskAwaiter = taskAwaiter2;
			int num = 0;
			if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_7fa3c0be0074475fb946e9d8136ace47 == 0)
			{
				num = 0;
			}
			switch (num)
			{
			}
		}
		return taskAwaiter.GetResult();
	}

	public async Task<ZResult> PostAsync(string name, int value)
	{
		TaskAwaiter<ZResult> taskAwaiter = PostAsync(name, twsXv6Pt0TqWJgLIYbQo.RyQPtm98V8J(ref value, twsXv6Pt0TqWJgLIYbQo.VJ5PtcFXjFw)).GetAwaiter();
		if (!taskAwaiter.IsCompleted)
		{
			int num = 0;
			if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_6851347940cd4439be5e0af07935dfb9 != 0)
			{
				num = 0;
			}
			switch (num)
			{
			}
			await taskAwaiter;
			TaskAwaiter<ZResult> taskAwaiter2 = default;
			taskAwaiter = taskAwaiter2;
		}
		return taskAwaiter.GetResult();
	}

	public async Task<ZResult> PostAsync(string name, long value)
	{
		TaskAwaiter<ZResult> taskAwaiter = PostAsync(name, em9AKIPstQFTyFbhg88p.RyQPtm98V8J(ref value, em9AKIPstQFTyFbhg88p.vjRPsssHvGb)).GetAwaiter();
		if (!taskAwaiter.IsCompleted)
		{
			await taskAwaiter;
			TaskAwaiter<ZResult> taskAwaiter2 = default;
			taskAwaiter = taskAwaiter2;
			int num = 0;
			if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_4f83509c06314609acf535386b4f0c3d != 0)
			{
				num = 0;
			}
			switch (num)
			{
			}
		}
		return taskAwaiter.GetResult();
	}

	public async Task<ZResult> PostAsync(string name, decimal value)
	{
		ZResult result = await PostAsync(name, JcM84iPorpsd2gl1G2jt.RyQPtm98V8J(ref value, JcM84iPorpsd2gl1G2jt.wf8PowAOB64));
		int num = 0;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_dd6240aeb46e481f98298eeaa3803307 != 0)
		{
			num = 0;
		}
		switch (num)
		{
		default:
			return result;
		case 0:
			return result;
		}
	}

	public async Task<ZResult> PostAsync(string name, string value)
	{
		NameValueCollection nameValueCollection = new NameValueCollection();
		bRCg2LPfjco89dAY95bx.RyQPtm98V8J(nameValueCollection, name, value, bRCg2LPfjco89dAY95bx.Fi4Pfbhtu5l);
		NameValueCollection nv = nameValueCollection;
		TaskAwaiter<ZResult> taskAwaiter = PostAsync(nv).GetAwaiter();
		int num = 0;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5e7209d6eaaa4fd4bfb5c9f2b2a4480c != 0)
		{
			num = 0;
		}
		switch (num)
		{
		default:
			if (!taskAwaiter.IsCompleted)
			{
				await taskAwaiter;
				TaskAwaiter<ZResult> taskAwaiter2 = default;
				taskAwaiter = taskAwaiter2;
			}
			return taskAwaiter.GetResult();
		}
	}

	public async Task<ZResult> PostAsync(byte[] data)
	{
		TaskAwaiter<ZResult> taskAwaiter = ConnectAsync(data).GetAwaiter();
		if (!taskAwaiter.IsCompleted)
		{
			await taskAwaiter;
			TaskAwaiter<ZResult> taskAwaiter2 = default;
			taskAwaiter = taskAwaiter2;
			int num = 0;
			if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_6047c4a60e7b4ce3bb5b0bd793ab5b8c == 0)
			{
				num = 0;
			}
			switch (num)
			{
			}
		}
		return taskAwaiter.GetResult();
	}

	[AsyncStateMachine(typeof(<PostAsync>d__103))]
	public Task<ZResult> PostAsync()
	{
		// ILSpy could not decompile this member: Error decompiling @060014E1 ZYing.ZClient.PostAsync: Index was out of range. Must be non-negative and less than the size of the collection. (Parameter 'index')
		// Please report the exception below,
		// along with the assembly it came from, at https://github.com/icsharpcode/ILSpy/issues/new
		// Error decompiling @060014E1 ZYing.ZClient.PostAsync
		// in assembly "C:/Users/ASUS/AppData/Local/Temp/zying-rea-20261115/named\ZYing.Work.dll"
		//  ---> System.ArgumentOutOfRangeException: Index was out of range. Must be non-negative and less than the size of the collection. (Parameter 'index')
		//    at System.ThrowHelper.ThrowArgumentOutOfRange_IndexMustBeLessException() + 0x1b
		//    at ICSharpCode.Decompiler.IL.ControlFlow.AsyncAwaitDecompiler.AnalyzeAwaitBlock(Block, ILVariable&, IField&, Int32&, Int32&) + 0x593
		//    at ICSharpCode.Decompiler.IL.ControlFlow.AsyncAwaitDecompiler.AnalyzeStateMachine(ILFunction) + 0x248
		//    at ICSharpCode.Decompiler.IL.ControlFlow.AsyncAwaitDecompiler.Run(ILFunction, ILTransformContext) + 0xdb
		//    at ICSharpCode.Decompiler.IL.ILFunction.RunTransforms(IEnumerable`1, ILTransformContext) + 0x75
		//    at ICSharpCode.Decompiler.CSharp.CSharpDecompiler.DecompileBody(IMethod, EntityDeclaration, DecompileRun, ITypeResolveContext, ExtensionInfo) + 0x3b9
		// -- continuing with outer exception (ICSharpCode.Decompiler.DecompilerException) --
		// 
	}

	static ZClient()
	{
		A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		A8RedEce1ZSTaHmUDS.vv9ejnCAD();
		R7c2WOBSNAiVDsbUZ4.u2ulMj8ByVd();
		UUID = McOijAPofgQxvrxvGV5s.RyQPtm98V8J(McOijAPofgQxvrxvGV5s.VvnPoChInDM);
		Items = new EItemCollection();
		Ticks = 1;
		byte[] array = new byte[9];
		jgALHnPwz0q20IWSX4hh.RyQPtm98V8J(array, __ldtoken(<PrivateImplementationDetails>.93AF66E3E627EA1C8EF0AA4822E6069B4D5438DC381F30F86991312C8DC84A5E), jgALHnPwz0q20IWSX4hh.mGmPMOsY0ou);
		_key = array;
		聚石塔北京 = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x64DDC4A4 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_dd6240aeb46e481f98298eeaa3803307);
		老吕 = new IPGroup(LIPTDTPfTJJkYYfdiLfx.RyQPtm98V8J(JsonConfig<AppConfig>.Current.HK ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2716E679 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_7df624928bfc47f38784b94f3192f82e) : A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0xD0E2915 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_c006ecad44af4f91bbb59ddc7246133d), LIPTDTPfTJJkYYfdiLfx.ICePfnfEbrA));
		int num = 1;
		if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_470e8e52815e40ad83294a93ad76b4aa != 0)
		{
			num = 1;
		}
		while (true)
		{
			switch (num)
			{
			case 2:
				刊登 = new IPGroup(LIPTDTPfTJJkYYfdiLfx.RyQPtm98V8J(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x45CB0B60 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5c4f0b48dde647e5b4b41ac0c211fb6c), LIPTDTPfTJJkYYfdiLfx.ICePfnfEbrA));
				香港 = new IPGroup(LIPTDTPfTJJkYYfdiLfx.RyQPtm98V8J((string)KAWovlu1GlK5dRK496VD(0x636DF021 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_c7210e071b3e401da22af49080b21ab3), LIPTDTPfTJJkYYfdiLfx.ICePfnfEbrA));
				择远 = new IPGroup(LIPTDTPfTJJkYYfdiLfx.RyQPtm98V8J(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4C24624 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_85d2b8790f40427d85672cf528d7a58e), LIPTDTPfTJJkYYfdiLfx.ICePfnfEbrA));
				速卖通 = new IPGroup(wTN8eoPf5OFMLVmgrrP5.RyQPtm98V8J(聚石塔北京, wTN8eoPf5OFMLVmgrrP5.FwePfUQJy3c));
				速卖通远程 = new IPGroup(wTN8eoPf5OFMLVmgrrP5.RyQPtm98V8J(聚石塔北京, wTN8eoPf5OFMLVmgrrP5.FwePfUQJy3c));
				美客多 = 网站;
				N11 = 网站;
				Closing = false;
				num = 0;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_96a450f6f5a94abc8729e962fe36bc75 == 0)
				{
					num = 0;
				}
				break;
			default:
			{
				dVt6qePEWaGkFNaR0cti.RyQPtm98V8J(Items, new EItem[2]
				{
					new EItem(null, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4E01305E ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_8431c63ac4f045acb30300812580a306), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3298F30F ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_eb2c289411da4cb6ba956a784b15c8f8), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x198572CF ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_725165346e184a6a9e1ca06eeca2c3c1)),
					new EItem(null, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x45CB0B60 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_5c4f0b48dde647e5b4b41ac0c211fb6c), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x502EDB35 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_6047c4a60e7b4ce3bb5b0bd793ab5b8c))
				}, dVt6qePEWaGkFNaR0cti.TRoPEEtlPKM);
				Server = new IPGroup(LIPTDTPfTJJkYYfdiLfx.RyQPtm98V8J(JsonConfig<AppConfig>.Current.HK ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6B87B174 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_264717b9aada4491a008eb41e1264844) : C5yeLrPfeCaEPp3uUNLk.RyQPtm98V8J(Items, C5yeLrPfeCaEPp3uUNLk.uLgPf9cLSU2), LIPTDTPfTJJkYYfdiLfx.ICePfnfEbrA));
				Thread thread = new Thread(CheckExpired);
				N0PNYXPfHl9dO2BZoIoK.RyQPtm98V8J(thread, true, N0PNYXPfHl9dO2BZoIoK.lJEPfiQaWS4);
				_thread = thread;
				y4xB4nPffpqF9I5RCor8.RyQPtm98V8J(_thread, y4xB4nPffpqF9I5RCor8.hS2PfCgTvOT);
				if (JsonConfig<AppConfig>.Current.LastLoginIsZeyuan)
				{
					zeyuan();
				}
				_current = (_root = new ZClient(Server));
				return;
			}
			case 1:
				网站 = new IPGroup(LIPTDTPfTJJkYYfdiLfx.RyQPtm98V8J(JsonConfig<AppConfig>.Current.HK ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4C2469C ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_85d2b8790f40427d85672cf528d7a58e) : A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7761DC26 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_2014a74f825b47e3b23645bee3edf31a), LIPTDTPfTJJkYYfdiLfx.ICePfnfEbrA));
				num = 2;
				break;
			}
		}
	}

	[Conditional("DEBUG")]
	private static void localhost()
	{
		IPAddress iPAddress = LIPTDTPfTJJkYYfdiLfx.RyQPtm98V8J(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0xC5717C7 ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_f6823ca2c2d44cdba0ce31084c91362a), LIPTDTPfTJJkYYfdiLfx.ICePfnfEbrA);
		Server = IPGroup.GetOrCreate(iPAddress);
		老吕 = (IPGroup)ar30PMu100i9SM0lolXd(iPAddress);
		刊登 = IPGroup.GetOrCreate(iPAddress);
		香港 = (IPGroup)ar30PMu100i9SM0lolXd(iPAddress);
		网站 = IPGroup.GetOrCreate(iPAddress);
		速卖通 = IPGroup.GetOrCreate(iPAddress);
		美客多 = IPGroup.GetOrCreate(iPAddress);
		N11 = IPGroup.GetOrCreate(iPAddress);
	}

	public static void zeyuan()
	{
		IPAddress iPAddress = LIPTDTPfTJJkYYfdiLfx.RyQPtm98V8J(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x30CD43DA ^ <Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_c139cf29f0154265ae58848394260b57), LIPTDTPfTJJkYYfdiLfx.ICePfnfEbrA);
		Server = IPGroup.GetOrCreate(iPAddress);
		老吕 = IPGroup.GetOrCreate(iPAddress);
		刊登 = IPGroup.GetOrCreate(iPAddress);
		香港 = IPGroup.GetOrCreate(iPAddress);
		网站 = IPGroup.GetOrCreate(iPAddress);
		速卖通 = IPGroup.GetOrCreate(iPAddress);
		美客多 = (IPGroup)ar30PMu100i9SM0lolXd(iPAddress);
	}

	private static ZClient CreateNew(IPGroup grp)
	{
		ZClient value = new ZClient(grp);
		Interlocked.Exchange(ref _current.Next, value);
		ZClient current = _current;
		Interlocked.Exchange(ref _current, _current.Next);
		return current;
	}

	public static void CheckExpired()
	{
		while (lrNWFjPfVRsA7UBQ1TAA.RyQPtm98V8J(_thread, lrNWFjPfVRsA7UBQ1TAA.ffLPfdmSX8c))
		{
			Expired(300000);
			int num = 0;
			while (num < 3000 && lrNWFjPfVRsA7UBQ1TAA.RyQPtm98V8J(_thread, lrNWFjPfVRsA7UBQ1TAA.ffLPfdmSX8c))
			{
				t7CL5mPf3aT5lpo8EA0m.RyQPtm98V8J(100, t7CL5mPf3aT5lpo8EA0m.RPZPfGeaR3t);
				num++;
				int num2 = 0;
				if (<Module>{9339e5e7-0cea-4616-8903-968ffb69e779}.m_9ba69d50f8ff4945865861f126c78db6.m_4a51f166446e46eba3624ab89b6366ef == 0)
				{
					num2 = 0;
				}
				switch (num2)
				{
				}
			}
		}
	}

	public static void CloseAll()
	{
		Closing = true;
		try
		{
			for (ZClient zClient = _root; zClient != null; zClient = zClient.Next)
			{
				if (zClient.IsActive)
				{
					zClient.Close();
				}
			}
		}
		finally
		{
			Closing = false;
		}
	}

	public static void Expired(int timeout)
	{
		DateTime dateTime = A0X3HkPtQh8HhvIRv9Y9.RyQPtm98V8J(A0X3HkPtQh8HhvIRv9Y9.GhQPtkN7xja);
		ZClient zClient = _root;
		while (zClient != null && lrNWFjPfVRsA7UBQ1TAA.RyQPtm98V8J(_thread, lrNWFjPfVRsA7UBQ1TAA.ffLPfdmSX8c))
		{
			try
			{
				if (oOlqHqu1cEGajxD6RJtj(zClient) && !zClient.Busy)
				{
					if (timeout == 0)
					{
						goto IL_0056;
					}
					TimeSpan timeSpan = sajliCP58rNuDdC0fXQB.RyQPtm98V8J(ref dateTime, zClient.ActiveTime, sajliCP58rNuDdC0fXQB.dB4P5YG8U8l);
					if (PqXddoPfXD6hAlGvjYAG.RyQPtm98V8J(ref timeSpan, PqXddoPfXD6hAlGvjYAG.CQMPfgCluIx) > (double)timeout)
					{
						goto IL_0056;
					}
				}
				continue;
				IL_0056:
				zClient.Close();
			}
			catch (Exception)
			{
			}
			zClient = zClient.Next;
		}
	}

	internal static bool kDju1cu1dAFeQX8gIOAJ()
	{
		return HQvdN4u1VHn6ornifLWZ == null;
	}

	internal static ZClient XgYvOTu13i6Xm8AdyZK8()
	{
		return HQvdN4u1VHn6ornifLWZ;
	}

	internal static object KAWovlu1GlK5dRK496VD(int P_0)
	{
		return A8RedEce1ZSTaHmUDS.DQmUrYZ9D(P_0);
	}

	internal static bool nigPZEu1X10BjV4U25bO(object P_0)
	{
		return ((ZResult)P_0).Done();
	}

	internal static void Yk9IQiu1gp7vJ0nH0CMS(object P_0, object P_1)
	{
		((IPGroup)P_0).Enqueue((ZClient)P_1);
	}

	internal static object ar30PMu100i9SM0lolXd(object P_0)
	{
		return IPGroup.GetOrCreate((IPAddress)P_0);
	}

	internal static bool oOlqHqu1cEGajxD6RJtj(object P_0)
	{
		return ((ZClient)P_0).IsActive;
	}
}
