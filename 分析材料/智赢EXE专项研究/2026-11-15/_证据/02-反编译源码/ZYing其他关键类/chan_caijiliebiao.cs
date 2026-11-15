using System;
using System.Collections.Specialized;
using System.ComponentModel;
using System.Diagnostics;
using System.Drawing;
using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;
using System.Text;
using System.Windows.Forms;
using Enumer;
using Hmg1hUwjXmJQ0hx3wd;
using Xmbn6iXVDuARF6WhpW;
using ZYing.Data;
using ZYing.Detail;
using ZYing.Dialog;
using ZYing.Interface;
using ZYing.Static;
using ZYing.UI;

namespace ZYing.Channel;

[ToolboxItem(false)]
public class 采集列表 : CBase
{
	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <Batch分类>d__31 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncVoidMethodBuilder <>t__builder;

		public 采集列表 <>4__this;

		public int localid;

		private ZClient <zc>5__2;

		private TaskAwaiter<ZResult> <>u__1;

		private static object IafXyqNb8XBlCgnMqkdr;

		private void MoveNext()
		{
			int num = <>1__state;
			采集列表 采集列表2 = <>4__this;
			try
			{
				TaskAwaiter<ZResult> awaiter;
				if (num != 0)
				{
					StringBuilder stringBuilder = new StringBuilder();
					int num2 = 0;
					if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b6fe9056d2c84afdb2f3799dd7895d25 != 0)
					{
						num2 = 0;
					}
					while (true)
					{
						switch (num2)
						{
						case 1:
							break;
						default:
						{
							for (int i = 0; i < jtwfoCIToQl9n5ohwkvu.qOys5bJ94(采集列表2.grid, jtwfoCIToQl9n5ohwkvu.CdqIT6mbZyE); i++)
							{
								EGSale eGSale = dPegXZITrT0v8bhnlqHy.qOys5bJ94(采集列表2.grid, i, dPegXZITrT0v8bhnlqHy.C3BITwTmsVc) as EGSale;
								if (eGSale.Selected)
								{
									yuX4loaHrxU06DUKTed.qOys5bJ94(stringBuilder, l6GFAG7hp5Ui898HhZ.qOys5bJ94((Cn1oVaaeOcrG1OdN6mm.qOys5bJ94(stringBuilder, Cn1oVaaeOcrG1OdN6mm.MVPa9bItHD) > 0) ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7C9144E6 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_c5ba8cdf0efd483d97721d48bfc441f0) : "", eGSale.Name, l6GFAG7hp5Ui898HhZ.iIvA7JOfD), yuX4loaHrxU06DUKTed.KVFaiRKWhU);
								}
							}
							<zc>5__2 = 采集列表2.Create();
							H2nI6Pcfnm36xOxZjj8.qOys5bJ94(<zc>5__2, 采集列表2.lblTotal, H2nI6Pcfnm36xOxZjj8.hIqcCZegBF);
							<zc>5__2.Command = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x5D6A6E0D ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_a038870306f14fbfb665c2e3a98b4d7b);
							num2 = 1;
							if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_79ef22e6a7c34985869c3fc9b18ef11b != 0)
							{
								num2 = 1;
							}
							continue;
						}
						}
						break;
					}
					NameValueCollection nameValueCollection = new NameValueCollection();
					w2YQcHHPZLmGrQrufBT.qOys5bJ94(nameValueCollection, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2E824E68 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_bf5be955d0984bfc94b8150139b53028), aGThQ0zBsC1uu6CBnB.qOys5bJ94(ref localid, aGThQ0zBsC1uu6CBnB.lvfIO2Ntg8), w2YQcHHPZLmGrQrufBT.quqHuAIY5J);
					w2YQcHHPZLmGrQrufBT.qOys5bJ94(nameValueCollection, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-(-71192797) ^ 0x3C216AB5 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9bf27d58a5584b2cb3d10730eccbd9c5), rXYd8UaVChLCL8Lkm5d.qOys5bJ94(stringBuilder, rXYd8UaVChLCL8Lkm5d.T7ZadVdu8C), w2YQcHHPZLmGrQrufBT.quqHuAIY5J);
					NameValueCollection nameValueCollection2 = nameValueCollection;
					awaiter = g4kg5vHKOnsqZfJULPF.qOys5bJ94(<zc>5__2, nameValueCollection2, g4kg5vHKOnsqZfJULPF.m7gHLw2plr).GetAwaiter();
					if (!awaiter.IsCompleted)
					{
						num = (<>1__state = 0);
						<>u__1 = awaiter;
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
				awaiter.GetResult();
				PWGKTQ28i8QujJOnnX4.qOys5bJ94(ehZmE0IwMoFJgYMdIUbr.qOys5bJ94(<zc>5__2, ehZmE0IwMoFJgYMdIUbr.jFEIwxgN0r2), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-1767913333 ^ -368168665 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_c5ba8cdf0efd483d97721d48bfc441f0), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
			}
			catch (Exception ex)
			{
				<>1__state = -2;
				<zc>5__2 = null;
				XpWRPcI0MQjGF59SwjbE.qOys5bJ94(ref <>t__builder, ex, XpWRPcI0MQjGF59SwjbE.GYcI0xetTsR);
				return;
			}
			<>1__state = -2;
			<zc>5__2 = null;
			rN63RpI0T5mLCF9FK9kW.qOys5bJ94(ref <>t__builder, rN63RpI0T5mLCF9FK9kW.xLWI0njWt8N);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			Px919kI052r2nu6WIGGs.qOys5bJ94(ref <>t__builder, stateMachine, Px919kI052r2nu6WIGGs.JJ6I0Uvdbui);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <Batch分类>d__31()
		{
			JRvNj5NbJAnGe8jxMYNm();
		}

		internal static bool EyvdRyNbYiB42C0SGWhF()
		{
			return IafXyqNb8XBlCgnMqkdr == null;
		}

		internal static object nSgiInNbSk0rFADTHe3e()
		{
			return IafXyqNb8XBlCgnMqkdr;
		}

		internal static void JRvNj5NbJAnGe8jxMYNm()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <Batch状态>d__30 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncVoidMethodBuilder <>t__builder;

		public string op;

		public 采集列表 <>4__this;

		private ZClient <zc>5__2;

		private TaskAwaiter<ZResult> <>u__1;

		private static object OBTRadNbziWYp3RaYOAr;

		private void MoveNext()
		{
			int num = <>1__state;
			采集列表 采集列表2 = <>4__this;
			try
			{
				int num2;
				if (num == 0)
				{
					num2 = 1;
					if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_fca4b5827c154beda800a45f1a8c082c != 0)
					{
						num2 = 1;
					}
					goto IL_0037;
				}
				if (!ccuqBrl774FBBcBTcHX.qOys5bJ94(op, (string)X7c4FtNo2j0HZkORsFTl(0x652128C9 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_8d18136f38d74e8f97d86f9cab9bb6ca), ccuqBrl774FBBcBTcHX.lpSlAEh9YT))
				{
					goto IL_0135;
				}
				goto end_IL_000d;
				IL_0037:
				TaskAwaiter<ZResult> awaiter;
				StringBuilder stringBuilder = default;
				switch (num2)
				{
				case 1:
					awaiter = <>u__1;
					<>u__1 = default;
					num = (<>1__state = -1);
					goto IL_010b;
				case 2:
					{
						NameValueCollection nameValueCollection = new NameValueCollection();
						w2YQcHHPZLmGrQrufBT.qOys5bJ94(nameValueCollection, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7F5EDAA4 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_71b15132f4c3474aaebfe1530c68df31), op, w2YQcHHPZLmGrQrufBT.quqHuAIY5J);
						w2YQcHHPZLmGrQrufBT.qOys5bJ94(nameValueCollection, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4F842E46 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_845968e715db46549494fad4e2bbf631), rXYd8UaVChLCL8Lkm5d.qOys5bJ94(stringBuilder, rXYd8UaVChLCL8Lkm5d.T7ZadVdu8C), w2YQcHHPZLmGrQrufBT.quqHuAIY5J);
						awaiter = g4kg5vHKOnsqZfJULPF.qOys5bJ94(<zc>5__2, nameValueCollection, g4kg5vHKOnsqZfJULPF.m7gHLw2plr).GetAwaiter();
						if (!awaiter.IsCompleted)
						{
							num = (<>1__state = 0);
							<>u__1 = awaiter;
							<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
							return;
						}
						goto IL_010b;
					}
					IL_010b:
					awaiter.GetResult();
					PWGKTQ28i8QujJOnnX4.qOys5bJ94(ehZmE0IwMoFJgYMdIUbr.qOys5bJ94(<zc>5__2, ehZmE0IwMoFJgYMdIUbr.jFEIwxgN0r2), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x33679C8E ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9eb1b575102d43aeb17e7aaeef3dc5dd), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
					goto end_IL_000d;
				}
				goto IL_0135;
				IL_0135:
				stringBuilder = new StringBuilder();
				for (int i = 0; i < jtwfoCIToQl9n5ohwkvu.qOys5bJ94(采集列表2.grid, jtwfoCIToQl9n5ohwkvu.CdqIT6mbZyE); i++)
				{
					EGSale eGSale = dPegXZITrT0v8bhnlqHy.qOys5bJ94(采集列表2.grid, i, dPegXZITrT0v8bhnlqHy.C3BITwTmsVc) as EGSale;
					if (HWAyw6Noyv1rQRFvlVD6(eGSale))
					{
						yuX4loaHrxU06DUKTed.qOys5bJ94(stringBuilder, l6GFAG7hp5Ui898HhZ.qOys5bJ94((Cn1oVaaeOcrG1OdN6mm.qOys5bJ94(stringBuilder, Cn1oVaaeOcrG1OdN6mm.MVPa9bItHD) > 0) ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4CD1E04B ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_282f1df2aaef433d8ead5f2d044ac539) : "", eGSale.Name, l6GFAG7hp5Ui898HhZ.iIvA7JOfD), yuX4loaHrxU06DUKTed.KVFaiRKWhU);
					}
				}
				<zc>5__2 = 采集列表2.Create();
				H2nI6Pcfnm36xOxZjj8.qOys5bJ94(<zc>5__2, 采集列表2.lblTotal, H2nI6Pcfnm36xOxZjj8.hIqcCZegBF);
				<zc>5__2.Command = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1D1171D1 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_8363360abdc34f81bce948eb97479daa);
				num2 = 2;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_01ceb1fa80cc4c7092dd54bea08b0961 == 0)
				{
					num2 = 2;
				}
				goto IL_0037;
				end_IL_000d:;
			}
			catch (Exception ex)
			{
				<>1__state = -2;
				<zc>5__2 = null;
				XpWRPcI0MQjGF59SwjbE.qOys5bJ94(ref <>t__builder, ex, XpWRPcI0MQjGF59SwjbE.GYcI0xetTsR);
				return;
			}
			<>1__state = -2;
			<zc>5__2 = null;
			rN63RpI0T5mLCF9FK9kW.qOys5bJ94(ref <>t__builder, rN63RpI0T5mLCF9FK9kW.xLWI0njWt8N);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			Px919kI052r2nu6WIGGs.qOys5bJ94(ref <>t__builder, stateMachine, Px919kI052r2nu6WIGGs.JJ6I0Uvdbui);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <Batch状态>d__30()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}

		internal static object X7c4FtNo2j0HZkORsFTl(int P_0)
		{
			return A8RedEce1ZSTaHmUDS.DQmUrYZ9D(P_0);
		}

		internal static bool HWAyw6Noyv1rQRFvlVD6(object P_0)
		{
			return ((EGSale)P_0).Selected;
		}

		internal static bool m9rFiINoO6p9NR4pulUf()
		{
			return OBTRadNbziWYp3RaYOAr == null;
		}

		internal static object PNCsQ4NoIRhfVHIbLq2k()
		{
			return OBTRadNbziWYp3RaYOAr;
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <Btn多选导入_Click>d__25 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncVoidMethodBuilder <>t__builder;

		public 采集列表 <>4__this;

		private ZClient <zc>5__2;

		private TaskAwaiter<ZResult> <>u__1;

		internal static object uxEAsJNohc20c8jqbvYb;

		private void MoveNext()
		{
			int num = <>1__state;
			采集列表 采集列表2 = <>4__this;
			try
			{
				int num2;
				if (num == 0)
				{
					num2 = 1;
					if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_25089bab62a74b43b37fc67ad5fef7a9 == 0)
					{
						num2 = 2;
					}
					goto IL_0037;
				}
				if (!PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(采集列表2.chkSelect, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0))
				{
					raZqQm15icvksyfKy1g.qOys5bJ94(采集列表2.chkSelect, true, raZqQm15icvksyfKy1g.hjx1UOoNFG);
					num2 = 0;
					if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_66964c0ee427455da09003aacb36f971 != 0)
					{
						num2 = 0;
					}
					goto IL_0037;
				}
				int num3 = 0;
				StringBuilder stringBuilder = new StringBuilder();
				for (int i = 0; i < jtwfoCIToQl9n5ohwkvu.qOys5bJ94(采集列表2.grid, jtwfoCIToQl9n5ohwkvu.CdqIT6mbZyE); i++)
				{
					EGSale eGSale = dPegXZITrT0v8bhnlqHy.qOys5bJ94(采集列表2.grid, i, dPegXZITrT0v8bhnlqHy.C3BITwTmsVc) as EGSale;
					if (eGSale.Selected)
					{
						yuX4loaHrxU06DUKTed.qOys5bJ94(stringBuilder, l6GFAG7hp5Ui898HhZ.qOys5bJ94((Cn1oVaaeOcrG1OdN6mm.qOys5bJ94(stringBuilder, Cn1oVaaeOcrG1OdN6mm.MVPa9bItHD) > 0) ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x5D6DD507 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_a038870306f14fbfb665c2e3a98b4d7b) : "", eGSale.Name, l6GFAG7hp5Ui898HhZ.iIvA7JOfD), yuX4loaHrxU06DUKTed.KVFaiRKWhU);
						num3++;
					}
				}
				if (num3 > 0)
				{
					goto IL_02ad;
				}
				PWGKTQ28i8QujJOnnX4.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1D117135 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_8363360abdc34f81bce948eb97479daa), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x689741B4 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ce1dfb55258f4152afc31779feefbcba), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
				goto end_IL_000d;
				IL_02ad:
				<zc>5__2 = 采集列表2.Create();
				gae4TFeKJZ2q8nKbBWh.qOys5bJ94(<zc>5__2, 采集列表2.lblTotal, gae4TFeKJZ2q8nKbBWh.TfYeLAJdSv);
				bbjfBeHyAJNM9yBL0JF.qOys5bJ94(<zc>5__2, 采集列表2.btn多选导入, bbjfBeHyAJNM9yBL0JF.HtTHh6cJOR);
				int num4 = 3;
				num2 = num4;
				goto IL_0037;
				IL_0037:
				TaskAwaiter<ZResult> awaiter;
				switch (num2)
				{
				default:
					goto end_IL_000d;
				case 0:
					goto end_IL_000d;
				case 2:
					awaiter = <>u__1;
					<>u__1 = default;
					num = (<>1__state = -1);
					goto IL_0320;
				case 1:
					break;
				case 3:
					{
						<zc>5__2.Command = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x78000307 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_488581fc99e442088402176de3de7cbd);
						awaiter = aHNK5Gwy20xoxfPG8WY.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x74E95DDE ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d59bc24a307e4bdca17ad3010cacc7b5), rXYd8UaVChLCL8Lkm5d.qOys5bJ94(stringBuilder, rXYd8UaVChLCL8Lkm5d.T7ZadVdu8C), aHNK5Gwy20xoxfPG8WY.Qk9wh71axH).GetAwaiter();
						if (!awaiter.IsCompleted)
						{
							num = (<>1__state = 0);
							<>u__1 = awaiter;
							<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
							return;
						}
						goto IL_0320;
					}
					IL_0320:
					awaiter.GetResult();
					PWGKTQ28i8QujJOnnX4.qOys5bJ94(ehZmE0IwMoFJgYMdIUbr.qOys5bJ94(<zc>5__2, ehZmE0IwMoFJgYMdIUbr.jFEIwxgN0r2), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4F844790 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_845968e715db46549494fad4e2bbf631), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
					if (!egJUDhIwrmqHmvKFHXe6.qOys5bJ94(<zc>5__2, egJUDhIwrmqHmvKFHXe6.o5mIwwv29e0))
					{
						采集列表2.doStat(采集列表2, EventArgs.Empty);
					}
					<zc>5__2 = null;
					goto end_IL_000d;
				}
				goto IL_02ad;
				end_IL_000d:;
			}
			catch (Exception ex)
			{
				<>1__state = -2;
				XpWRPcI0MQjGF59SwjbE.qOys5bJ94(ref <>t__builder, ex, XpWRPcI0MQjGF59SwjbE.GYcI0xetTsR);
				return;
			}
			<>1__state = -2;
			rN63RpI0T5mLCF9FK9kW.qOys5bJ94(ref <>t__builder, rN63RpI0T5mLCF9FK9kW.xLWI0njWt8N);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			Px919kI052r2nu6WIGGs.qOys5bJ94(ref <>t__builder, stateMachine, Px919kI052r2nu6WIGGs.JJ6I0Uvdbui);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <Btn多选导入_Click>d__25()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}

		internal static bool OE50AYNoaUt9vyMRQBxb()
		{
			return uxEAsJNohc20c8jqbvYb == null;
		}

		internal static object kBR9rBNoNa9JoFhQF8C8()
		{
			return uxEAsJNohc20c8jqbvYb;
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <btn导入主库_Click>d__32 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncVoidMethodBuilder <>t__builder;

		public 采集列表 <>4__this;

		private ZClient <zc>5__2;

		private TaskAwaiter<ZResult> <>u__1;

		internal static object scQdESNoPZa0wlqVrV3q;

		private void MoveNext()
		{
			int num = <>1__state;
			采集列表 采集列表2 = <>4__this;
			try
			{
				if (num == 0)
				{
					goto IL_016f;
				}
				TaskAwaiter<ZResult> awaiter;
				if (!VuIOJhwICsBsTi41HAf.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4F8B101C ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b480309f51bd48c194e1ed3315e9d650), VuIOJhwICsBsTi41HAf.qb4w2dRP3O))
				{
					<zc>5__2 = 采集列表2.Create();
					gae4TFeKJZ2q8nKbBWh.qOys5bJ94(<zc>5__2, 采集列表2.lblTotal, gae4TFeKJZ2q8nKbBWh.TfYeLAJdSv);
					<zc>5__2.Command = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x303377CD ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_7a1fed405f4748f2b7b141bf9600fe2e);
					awaiter = prOfdaV8ObL7SgZkVMD.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x328F3638 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_1599679ee1724b48bf7a40ba1cf1dff5), 0, prOfdaV8ObL7SgZkVMD.HtrVYttFJU).GetAwaiter();
					if (!awaiter.IsCompleted)
					{
						num = (<>1__state = 0);
						<>u__1 = awaiter;
						<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
						return;
					}
					goto IL_0127;
				}
				任务侵权检查.Instance.PType = 2;
				qXuYODFTYbD2eHsJK31.qOys5bJ94(任务.Instance, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2EE74908 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0f5d594c7dcb4a7f9e6cc9d946db1eec), qXuYODFTYbD2eHsJK31.y15Fn7pQ5t);
				goto end_IL_001a;
				IL_0127:
				awaiter.GetResult();
				int num2 = 0;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_603178892e9e413b98f3e6aa3da54b45 == 0)
				{
					num2 = 0;
				}
				switch (num2)
				{
				case 1:
					break;
				default:
					qcnWREIKQ4qeRugsTPyG.qOys5bJ94(<zc>5__2, qcnWREIKQ4qeRugsTPyG.kWeIKkM7Ss3);
					PWGKTQ28i8QujJOnnX4.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x11F8173C ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9cd5a0aaeba04bb4b081b4b5fe90a190), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x5080178D ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_fca4b5827c154beda800a45f1a8c082c), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
					采集列表2.Load();
					goto end_IL_001a;
				}
				goto IL_016f;
				IL_016f:
				awaiter = <>u__1;
				<>u__1 = default;
				num = (<>1__state = -1);
				goto IL_0127;
				end_IL_001a:;
			}
			catch (Exception ex)
			{
				<>1__state = -2;
				<zc>5__2 = null;
				XpWRPcI0MQjGF59SwjbE.qOys5bJ94(ref <>t__builder, ex, XpWRPcI0MQjGF59SwjbE.GYcI0xetTsR);
				return;
			}
			<>1__state = -2;
			<zc>5__2 = null;
			rN63RpI0T5mLCF9FK9kW.qOys5bJ94(ref <>t__builder, rN63RpI0T5mLCF9FK9kW.xLWI0njWt8N);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			Px919kI052r2nu6WIGGs.qOys5bJ94(ref <>t__builder, stateMachine, Px919kI052r2nu6WIGGs.JJ6I0Uvdbui);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <btn导入主库_Click>d__32()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}

		internal static bool gFpdYRNouKMbMqGJTXHs()
		{
			return scQdESNoPZa0wlqVrV3q == null;
		}

		internal static object e9kqZMNoK0oYu473KuI1()
		{
			return scQdESNoPZa0wlqVrV3q;
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <doCommand>d__22 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncVoidMethodBuilder <>t__builder;

		public 采集列表 <>4__this;

		public string cmd;

		public Para p;

		private ZClient <zc>5__2;

		private TaskAwaiter<ZResult> <>u__1;

		internal static object fYNYPDNoLjKBTvFaCc34;

		private void MoveNext()
		{
			int num = <>1__state;
			采集列表 采集列表2 = <>4__this;
			try
			{
				TaskAwaiter<ZResult> awaiter;
				if (num != 0)
				{
					<zc>5__2 = 采集列表2.Create();
					H2nI6Pcfnm36xOxZjj8.qOys5bJ94(<zc>5__2, 采集列表2.lblTotal, H2nI6Pcfnm36xOxZjj8.hIqcCZegBF);
					<zc>5__2.Command = l6GFAG7hp5Ui898HhZ.qOys5bJ94((string)A8WrGPNo4nOEs90wER9H(0x1D10ACB1 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_8363360abdc34f81bce948eb97479daa), cmd, l6GFAG7hp5Ui898HhZ.iIvA7JOfD);
					awaiter = pLOtdEVll8GQiRmMVYf.qOys5bJ94(<zc>5__2, p, pLOtdEVll8GQiRmMVYf.uE1VFgR5IC).GetAwaiter();
					if (!awaiter.IsCompleted)
					{
						num = (<>1__state = 0);
						<>u__1 = awaiter;
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
				awaiter.GetResult();
				if (ij7l1jIaTgbitH4R9qXA.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x508177B3 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_fca4b5827c154beda800a45f1a8c082c), ij7l1jIaTgbitH4R9qXA.oYRIanWr2Cy))
				{
					RAMjtK2KM6Bqxnul2X0I.qOys5bJ94(KindLocal.采集分类, bkkF38zeN4Ui4actsbL.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2D24B125 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ae94be659ed84b8b8b4b11eef55c3fd4), bkkF38zeN4Ui4actsbL.q8Tz9IXwE6), RAMjtK2KM6Bqxnul2X0I.GaV2Kx4HNEb);
				}
				if (ij7l1jIaTgbitH4R9qXA.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x26BEBF55 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_e6be942033a74c0baffa094b080e9e6b), ij7l1jIaTgbitH4R9qXA.oYRIanWr2Cy))
				{
					hFr3Y02LWAwx1dFWpS5n.qOys5bJ94(采集列表2.PSource, bkkF38zeN4Ui4actsbL.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x11FF6140 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9cd5a0aaeba04bb4b081b4b5fe90a190), bkkF38zeN4Ui4actsbL.q8Tz9IXwE6), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3A4773CA ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_c0fff80a6c6a4ab1825b648985b91c3a), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2D248585 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ae94be659ed84b8b8b4b11eef55c3fd4), (string)A8WrGPNo4nOEs90wER9H(0x48544151 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9e7c8ea59e0745258e93689531c631f8), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2D25D2D7 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ae94be659ed84b8b8b4b11eef55c3fd4), hFr3Y02LWAwx1dFWpS5n.xj52LEKYjv7);
				}
				if (!ij7l1jIaTgbitH4R9qXA.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4F8C7476 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b480309f51bd48c194e1ed3315e9d650), ij7l1jIaTgbitH4R9qXA.oYRIanWr2Cy))
				{
					goto IL_0529;
				}
				Record record = bkkF38zeN4Ui4actsbL.qOys5bJ94(<zc>5__2, (string)A8WrGPNo4nOEs90wER9H(0x3367FE52 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9eb1b575102d43aeb17e7aaeef3dc5dd), bkkF38zeN4Ui4actsbL.q8Tz9IXwE6);
				goto IL_0732;
				IL_0732:
				if (g7YrDea5RMpcq824Kup.qOys5bJ94(record, g7YrDea5RMpcq824Kup.KdwaUFxfi9))
				{
					goto IL_0468;
				}
				IVtS4esKnVCkgWprQeG.qOys5bJ94(采集列表2.PStat, record, false, IVtS4esKnVCkgWprQeG.h2rsLEv9Ua);
				goto IL_0529;
				IL_0468:
				int num2;
				if (UDmp02IPfs6QxhZxW4f.qOys5bJ94(record, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x18072807 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_fcd2f31ec7e946f5991ef1cf55616ab6), UDmp02IPfs6QxhZxW4f.gY2IuTibQO) == 3000)
				{
					采集列表2._待审核 = UDmp02IPfs6QxhZxW4f.qOys5bJ94(record, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x19F4DD0C ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_de4c858d184a45e7a9ea91377a1f9cba), UDmp02IPfs6QxhZxW4f.gY2IuTibQO);
					num2 = 0;
					if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b912e71988d0440f99d561d4df651c0e != 0)
					{
						num2 = 1;
					}
					goto IL_002a;
				}
				goto IL_0732;
				IL_04db:
				Record record2 = bkkF38zeN4Ui4actsbL.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6526B533 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_8d18136f38d74e8f97d86f9cab9bb6ca), bkkF38zeN4Ui4actsbL.q8Tz9IXwE6);
				ELoader lblTotal = 采集列表2.lblTotal;
				string text = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6894F3AA ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ce1dfb55258f4152afc31779feefbcba);
				int num3 = B2INt4zHoYYxfKI6bZ2.qOys5bJ94(record2, B2INt4zHoYYxfKI6bZ2.PWJzitpeBx);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(lblTotal, MaKBHa0rb8pCci0Q1b.qOys5bJ94(text, aGThQ0zBsC1uu6CBnB.qOys5bJ94(ref num3, aGThQ0zBsC1uu6CBnB.lvfIO2Ntg8), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x41B69144 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_6759e61c404645aea927fcf2b16eb7a6), MaKBHa0rb8pCci0Q1b.UNjcxHPJw), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				IVtS4esKnVCkgWprQeG.qOys5bJ94(采集列表2.PTurn, record2, false, IVtS4esKnVCkgWprQeG.h2rsLEv9Ua);
				DhXuoBInr9ogqFZdurTR.qOys5bJ94(采集列表2.grid, record2, DhXuoBInr9ogqFZdurTR.cjFInwAUKkk);
				if (PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(采集列表2.chkSelect, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0))
				{
					Qw9jC1ITHWsCwqvA3oNj.qOys5bJ94(采集列表2.grid, true, Qw9jC1ITHWsCwqvA3oNj.er2ITiRaVXc);
				}
				goto end_IL_001a;
				IL_0314:
				if (ij7l1jIaTgbitH4R9qXA.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x303472E3 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_7a1fed405f4748f2b7b141bf9600fe2e), ij7l1jIaTgbitH4R9qXA.oYRIanWr2Cy))
				{
					IVtS4esKnVCkgWprQeG.qOys5bJ94(采集列表2.PType, bkkF38zeN4Ui4actsbL.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4CD15EDB ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_282f1df2aaef433d8ead5f2d044ac539), bkkF38zeN4Ui4actsbL.q8Tz9IXwE6), false, IVtS4esKnVCkgWprQeG.h2rsLEv9Ua);
					num2 = 3;
					if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_a2336f0ce5734370adca67289ea74f9e != 0)
					{
						num2 = 2;
					}
					goto IL_002a;
				}
				goto IL_04db;
				IL_0529:
				if (ij7l1jIaTgbitH4R9qXA.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2D25B2E1 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ae94be659ed84b8b8b4b11eef55c3fd4), ij7l1jIaTgbitH4R9qXA.oYRIanWr2Cy))
				{
					IVtS4esKnVCkgWprQeG.qOys5bJ94(采集列表2.PFlag, bkkF38zeN4Ui4actsbL.qOys5bJ94(<zc>5__2, (string)A8WrGPNo4nOEs90wER9H(0x40E705ED ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_943c194482ce4adf8f2a36eb84b85a8d), bkkF38zeN4Ui4actsbL.q8Tz9IXwE6), false, IVtS4esKnVCkgWprQeG.h2rsLEv9Ua);
				}
				if (ij7l1jIaTgbitH4R9qXA.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x173344F ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_dfe08b55d34746fe8ba0141c88660b08), ij7l1jIaTgbitH4R9qXA.oYRIanWr2Cy))
				{
					IVtS4esKnVCkgWprQeG.qOys5bJ94(采集列表2.PLevel, bkkF38zeN4Ui4actsbL.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2E86CFB6 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_bf5be955d0984bfc94b8150139b53028), bkkF38zeN4Ui4actsbL.q8Tz9IXwE6), false, IVtS4esKnVCkgWprQeG.h2rsLEv9Ua);
				}
				if (ij7l1jIaTgbitH4R9qXA.qOys5bJ94(<zc>5__2, (string)A8WrGPNo4nOEs90wER9H(0x342DAC49 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_450309df683f4e4d8ed93c0e244cef4f), ij7l1jIaTgbitH4R9qXA.oYRIanWr2Cy))
				{
					IVtS4esKnVCkgWprQeG.qOys5bJ94(采集列表2.PGrade, bkkF38zeN4Ui4actsbL.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x41B5DE50 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_6759e61c404645aea927fcf2b16eb7a6), bkkF38zeN4Ui4actsbL.q8Tz9IXwE6), false, IVtS4esKnVCkgWprQeG.h2rsLEv9Ua);
					goto IL_0314;
				}
				num2 = 1;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_a2336f0ce5734370adca67289ea74f9e == 0)
				{
					num2 = 2;
				}
				goto IL_002a;
				IL_002a:
				switch (num2)
				{
				case 2:
					break;
				default:
					goto IL_0468;
				case 3:
					goto IL_04db;
				case 1:
					goto IL_0732;
				}
				goto IL_0314;
				end_IL_001a:;
			}
			catch (Exception ex)
			{
				<>1__state = -2;
				<zc>5__2 = null;
				XpWRPcI0MQjGF59SwjbE.qOys5bJ94(ref <>t__builder, ex, XpWRPcI0MQjGF59SwjbE.GYcI0xetTsR);
				return;
			}
			<>1__state = -2;
			<zc>5__2 = null;
			rN63RpI0T5mLCF9FK9kW.qOys5bJ94(ref <>t__builder, rN63RpI0T5mLCF9FK9kW.xLWI0njWt8N);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			Px919kI052r2nu6WIGGs.qOys5bJ94(ref <>t__builder, stateMachine, Px919kI052r2nu6WIGGs.JJ6I0Uvdbui);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <doCommand>d__22()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}

		internal static object A8WrGPNo4nOEs90wER9H(int P_0)
		{
			return A8RedEce1ZSTaHmUDS.DQmUrYZ9D(P_0);
		}

		internal static bool igjOwMNol62rO5asT4YS()
		{
			return fYNYPDNoLjKBTvFaCc34 == null;
		}

		internal static object aaRL0lNoFlFa3LFPb6KG()
		{
			return fYNYPDNoLjKBTvFaCc34;
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <产品删除>d__28 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncVoidMethodBuilder <>t__builder;

		public 采集列表 <>4__this;

		private ZClient <zc>5__2;

		private TaskAwaiter<ZResult> <>u__1;

		private static object fLSQkPNomEBnElyDd5Fa;

		private void MoveNext()
		{
			int num = <>1__state;
			采集列表 采集列表2 = <>4__this;
			try
			{
				TaskAwaiter<ZResult> awaiter = default;
				int num2;
				if (num == 0)
				{
					awaiter = <>u__1;
					<>u__1 = default;
					num2 = 2;
					if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_1dfe6a356d324a4aba1c21d85b6856ea == 0)
					{
						num2 = 2;
					}
					goto IL_001d;
				}
				string text = default;
				if (PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(采集列表2.chkSelect, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0))
				{
					StringBuilder stringBuilder = new StringBuilder();
					for (int i = 0; i < jtwfoCIToQl9n5ohwkvu.qOys5bJ94(采集列表2.grid, jtwfoCIToQl9n5ohwkvu.CdqIT6mbZyE); i++)
					{
						EGSale eGSale = dPegXZITrT0v8bhnlqHy.qOys5bJ94(采集列表2.grid, i, dPegXZITrT0v8bhnlqHy.C3BITwTmsVc) as EGSale;
						if (eGSale.Selected)
						{
							yuX4loaHrxU06DUKTed.qOys5bJ94(stringBuilder, l6GFAG7hp5Ui898HhZ.qOys5bJ94((Cn1oVaaeOcrG1OdN6mm.qOys5bJ94(stringBuilder, Cn1oVaaeOcrG1OdN6mm.MVPa9bItHD) > 0) ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7C9144E6 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_c5ba8cdf0efd483d97721d48bfc441f0) : "", eGSale.Name, l6GFAG7hp5Ui898HhZ.iIvA7JOfD), yuX4loaHrxU06DUKTed.KVFaiRKWhU);
						}
					}
					if (Cn1oVaaeOcrG1OdN6mm.qOys5bJ94(stringBuilder, Cn1oVaaeOcrG1OdN6mm.MVPa9bItHD) == 0)
					{
						PWGKTQ28i8QujJOnnX4.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4851EFB7 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9e7c8ea59e0745258e93689531c631f8), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x26BECE3F ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_e6be942033a74c0baffa094b080e9e6b), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
						num2 = 0;
						if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_21cb4653d1d542f1b694f9137faa29e1 != 0)
						{
							num2 = 0;
						}
						goto IL_001d;
					}
					text = rXYd8UaVChLCL8Lkm5d.qOys5bJ94(stringBuilder, rXYd8UaVChLCL8Lkm5d.T7ZadVdu8C);
					int value = w181pFiayTLKYtievMQ.qOys5bJ94(text, w181pFiayTLKYtievMQ.wlMiNg2H70).Length;
					DefaultInterpolatedStringHandler defaultInterpolatedStringHandler = new DefaultInterpolatedStringHandler(18, 1);
					NYvZ1aeRRY049ryT5q.qOys5bJ94(ref defaultInterpolatedStringHandler, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2D230A09 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ae94be659ed84b8b8b4b11eef55c3fd4), NYvZ1aeRRY049ryT5q.XJe9c0mjP);
					defaultInterpolatedStringHandler.AppendFormatted(value);
					NYvZ1aeRRY049ryT5q.qOys5bJ94(ref defaultInterpolatedStringHandler, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4F82231E ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_845968e715db46549494fad4e2bbf631), NYvZ1aeRRY049ryT5q.XJe9c0mjP);
					if (VuIOJhwICsBsTi41HAf.qOys5bJ94(GldZhAHNZqVdNwcyfr.qOys5bJ94(ref defaultInterpolatedStringHandler, GldZhAHNZqVdNwcyfr.wQBi03s4Z), VuIOJhwICsBsTi41HAf.qb4w2dRP3O))
					{
						<zc>5__2 = 采集列表2.Create();
						num2 = 1;
						if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_53af840f9e7849caa56dac3b5b1de46f != 0)
						{
							num2 = 0;
						}
						goto IL_001d;
					}
				}
				else
				{
					raZqQm15icvksyfKy1g.qOys5bJ94(采集列表2.chkSelect, true, raZqQm15icvksyfKy1g.hjx1UOoNFG);
				}
				goto end_IL_000d;
				IL_001d:
				while (true)
				{
					switch (num2)
					{
					case 3:
						PWGKTQ28i8QujJOnnX4.qOys5bJ94(ehZmE0IwMoFJgYMdIUbr.qOys5bJ94(<zc>5__2, ehZmE0IwMoFJgYMdIUbr.jFEIwxgN0r2), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x5D6D684D ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_a038870306f14fbfb665c2e3a98b4d7b), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
						<zc>5__2 = null;
						break;
					case 1:
						H2nI6Pcfnm36xOxZjj8.qOys5bJ94(<zc>5__2, 采集列表2.lblTotal, H2nI6Pcfnm36xOxZjj8.hIqcCZegBF);
						<zc>5__2.Command = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-1767913333 ^ -312211550 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ea7ba9c275744d9da2d713459cb8f164);
						awaiter = aHNK5Gwy20xoxfPG8WY.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x581F0599 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ed88ade44dea498194ae1c8221333918), text, aHNK5Gwy20xoxfPG8WY.Qk9wh71axH).GetAwaiter();
						if (awaiter.IsCompleted)
						{
							goto IL_01e8;
						}
						num = (<>1__state = 0);
						<>u__1 = awaiter;
						<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
						return;
					case 2:
						num = (<>1__state = -1);
						goto IL_01e8;
					case 0:
						break;
					}
					break;
					IL_01e8:
					awaiter.GetResult();
					采集列表2.doFresh(采集列表2, EventArgs.Empty);
					num2 = 3;
				}
				end_IL_000d:;
			}
			catch (Exception ex)
			{
				<>1__state = -2;
				XpWRPcI0MQjGF59SwjbE.qOys5bJ94(ref <>t__builder, ex, XpWRPcI0MQjGF59SwjbE.GYcI0xetTsR);
				return;
			}
			<>1__state = -2;
			rN63RpI0T5mLCF9FK9kW.qOys5bJ94(ref <>t__builder, rN63RpI0T5mLCF9FK9kW.xLWI0njWt8N);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			Px919kI052r2nu6WIGGs.qOys5bJ94(ref <>t__builder, stateMachine, Px919kI052r2nu6WIGGs.JJ6I0Uvdbui);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <产品删除>d__28()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}

		internal static bool sCChn9NoQuGGXrcRk5bF()
		{
			return fLSQkPNomEBnElyDd5Fa == null;
		}

		internal static object Ov9A5pNok2kv1oKRbZeC()
		{
			return fLSQkPNomEBnElyDd5Fa;
		}
	}

	[StructLayout(LayoutKind.Auto)]
	[CompilerGenerated]
	private struct <产品合并>d__24 : IAsyncStateMachine
	{
		public int <>1__state;

		public AsyncVoidMethodBuilder <>t__builder;

		public 采集列表 <>4__this;

		private ZClient <zc>5__2;

		private TaskAwaiter<ZResult> <>u__1;

		internal static object fqMZXVNoZ21WnS3yHfpf;

		private void MoveNext()
		{
			int num = <>1__state;
			采集列表 采集列表2 = <>4__this;
			try
			{
				TaskAwaiter<ZResult> awaiter;
				if (num == 0)
				{
					awaiter = <>u__1;
					<>u__1 = default;
					num = (<>1__state = -1);
					goto IL_0390;
				}
				StringBuilder stringBuilder = default;
				int num3;
				if (!PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(采集列表2.chkSelect, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0))
				{
					raZqQm15icvksyfKy1g.qOys5bJ94(采集列表2.chkSelect, true, raZqQm15icvksyfKy1g.hjx1UOoNFG);
				}
				else
				{
					int num2 = 0;
					stringBuilder = new StringBuilder();
					for (int i = 0; i < jtwfoCIToQl9n5ohwkvu.qOys5bJ94(采集列表2.grid, jtwfoCIToQl9n5ohwkvu.CdqIT6mbZyE); i++)
					{
						EGSale eGSale = dPegXZITrT0v8bhnlqHy.qOys5bJ94(采集列表2.grid, i, dPegXZITrT0v8bhnlqHy.C3BITwTmsVc) as EGSale;
						if (eGSale.Selected)
						{
							yuX4loaHrxU06DUKTed.qOys5bJ94(stringBuilder, l6GFAG7hp5Ui898HhZ.qOys5bJ94((Cn1oVaaeOcrG1OdN6mm.qOys5bJ94(stringBuilder, Cn1oVaaeOcrG1OdN6mm.MVPa9bItHD) > 0) ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2E826F50 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_bf5be955d0984bfc94b8150139b53028) : "", eGSale.Name, l6GFAG7hp5Ui898HhZ.iIvA7JOfD), yuX4loaHrxU06DUKTed.KVFaiRKWhU);
							num2++;
						}
					}
					if (num2 > 1)
					{
						if (VuIOJhwICsBsTi41HAf.qOys5bJ94(MaKBHa0rb8pCci0Q1b.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x65DD2964 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_7d77dc0ef27547e4ab0c2f1b6c3778ad), aGThQ0zBsC1uu6CBnB.qOys5bJ94(ref num2, aGThQ0zBsC1uu6CBnB.lvfIO2Ntg8), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x36B777D5 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_12046703b6504e42927f2daf3fea0cdd), MaKBHa0rb8pCci0Q1b.UNjcxHPJw), VuIOJhwICsBsTi41HAf.qb4w2dRP3O))
						{
							<zc>5__2 = 采集列表2.Create();
							num3 = 1;
							if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b227bbab40e0440284ca7e684c696e92 == 0)
							{
								num3 = 1;
							}
							goto IL_005a;
						}
					}
					else
					{
						PWGKTQ28i8QujJOnnX4.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x36B77401 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_12046703b6504e42927f2daf3fea0cdd), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2951394 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_21cb4653d1d542f1b694f9137faa29e1), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
					}
				}
				goto end_IL_004a;
				IL_0118:
				<zc>5__2 = null;
				goto end_IL_004a;
				IL_005a:
				while (true)
				{
					switch (num3)
					{
					case 3:
						break;
					case 1:
						H2nI6Pcfnm36xOxZjj8.qOys5bJ94(<zc>5__2, 采集列表2.lblTotal, H2nI6Pcfnm36xOxZjj8.hIqcCZegBF);
						num3 = 0;
						if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_81e2a82cc3e14fb9af28d1287da33791 != 0)
						{
							num3 = 0;
						}
						continue;
					case 2:
						goto IL_03ba;
					default:
						bbjfBeHyAJNM9yBL0JF.qOys5bJ94(<zc>5__2, 采集列表2.btn状态, bbjfBeHyAJNM9yBL0JF.HtTHh6cJOR);
						<zc>5__2.Command = A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x48539E97 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9e7c8ea59e0745258e93689531c631f8);
						num3 = 3;
						continue;
					}
					break;
				}
				awaiter = aHNK5Gwy20xoxfPG8WY.qOys5bJ94(<zc>5__2, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x69170A39 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_88a9b435230c4eaf827b56a79d6ee53c), rXYd8UaVChLCL8Lkm5d.qOys5bJ94(stringBuilder, rXYd8UaVChLCL8Lkm5d.T7ZadVdu8C), aHNK5Gwy20xoxfPG8WY.Qk9wh71axH).GetAwaiter();
				if (!awaiter.IsCompleted)
				{
					num = (<>1__state = 0);
					<>u__1 = awaiter;
					<>t__builder.AwaitUnsafeOnCompleted(ref awaiter, ref this);
					return;
				}
				goto IL_0390;
				IL_0390:
				awaiter.GetResult();
				PWGKTQ28i8QujJOnnX4.qOys5bJ94(ehZmE0IwMoFJgYMdIUbr.qOys5bJ94(<zc>5__2, ehZmE0IwMoFJgYMdIUbr.jFEIwxgN0r2), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x5240094C ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_10ffe2be534f4959b052b27cc6e82b39), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
				if (!egJUDhIwrmqHmvKFHXe6.qOys5bJ94(<zc>5__2, egJUDhIwrmqHmvKFHXe6.o5mIwwv29e0))
				{
					num3 = 2;
					goto IL_005a;
				}
				goto IL_0118;
				IL_03ba:
				采集列表2.doStat(采集列表2, EventArgs.Empty);
				goto IL_0118;
				end_IL_004a:;
			}
			catch (Exception ex)
			{
				<>1__state = -2;
				XpWRPcI0MQjGF59SwjbE.qOys5bJ94(ref <>t__builder, ex, XpWRPcI0MQjGF59SwjbE.GYcI0xetTsR);
				return;
			}
			<>1__state = -2;
			rN63RpI0T5mLCF9FK9kW.qOys5bJ94(ref <>t__builder, rN63RpI0T5mLCF9FK9kW.xLWI0njWt8N);
		}

		void IAsyncStateMachine.MoveNext()
		{
			//ILSpy generated this explicit interface implementation from .override directive in MoveNext
			this.MoveNext();
		}

		[DebuggerHidden]
		private void SetStateMachine(IAsyncStateMachine stateMachine)
		{
			Px919kI052r2nu6WIGGs.qOys5bJ94(ref <>t__builder, stateMachine, Px919kI052r2nu6WIGGs.JJ6I0Uvdbui);
		}

		void IAsyncStateMachine.SetStateMachine(IAsyncStateMachine stateMachine)
		{
			//ILSpy generated this explicit interface implementation from .override directive in SetStateMachine
			this.SetStateMachine(stateMachine);
		}

		static <产品合并>d__24()
		{
			A8RedEce1ZSTaHmUDS.iR3aCLEY7();
		}

		internal static bool biL43lNoBuFfxdvYvYKW()
		{
			return fqMZXVNoZ21WnS3yHfpf == null;
		}

		internal static object RiwnBvNoqk77X8ivkgbj()
		{
			return fqMZXVNoZ21WnS3yHfpf;
		}
	}

	public const string Key = "SList";

	private int _待审核;

	private IContainer components;

	private EPanelEx right;

	private ELayer main;

	private EGallery grid;

	private EFilter PType;

	private EFilter PGrade;

	private EFilter PStat;

	private ESearch PWord;

	private ELoader lblTotal;

	private ETurn PTurn;

	private ELocalSearch PKind;

	private ELine eLine1;

	private EFilter PFlag;

	private ECheck PBan;

	private EButton btnGO;

	private EText PKey;

	private ECheck chkSelect;

	private ELabel lblAll;

	private EButton btn状态;

	private ECombo PLogin;

	private ESelectDateRange PDate;

	private ETextPair Price;

	private ESelect selOP;

	private ELayerAuto lay多选;

	private EButton btn分类;

	private EButton btn删除;

	private EFilter PLevel;

	private ESelect PSize;

	private EButton btn导入主库;

	private ESelect PArea;

	private EDropRootMidLinkageSearch PSource;

	private EButton btn多选导入;

	private ETextPair txtDelays;

	internal static 采集列表 PWyTwPyrM2qPxaMNINe3;

	public static 采集列表 Instance => Global.Create(pGlMXC230tpo5NElPvg.qOys5bJ94(typeof(采集列表).TypeHandle, pGlMXC230tpo5NElPvg.uhF2G1N8xs)) as 采集列表;

	public static void MyFocus()
	{
		if (txqwipI5KqIYBCaK82PF.qOys5bJ94(MainForm.Instance, txqwipI5KqIYBCaK82PF.RgJI5Lyc8uD) == FormWindowState.Minimized)
		{
			hSPNtII5l7dbknIeM3go.qOys5bJ94(MainForm.Instance, FormWindowState.Normal, hSPNtII5l7dbknIeM3go.IuXI5FZIqRn);
		}
		qXuYODFTYbD2eHsJK31.qOys5bJ94(产品.Instance, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x689705DA ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ce1dfb55258f4152afc31779feefbcba), qXuYODFTYbD2eHsJK31.y15Fn7pQ5t);
	}

	public 采集列表()
	{
		A8RedEce1ZSTaHmUDS.vv9ejnCAD();
		R7c2WOBSNAiVDsbUZ4.bjoPvRdTrxh();
		base..ctor();
		InitializeComponent();
		Sh3ESJIT7ZuAe1EoPlf5.qOys5bJ94(this, whole_SizeChanged, Sh3ESJIT7ZuAe1EoPlf5.n7KITAPB7Es);
		zFdO2SyrnGnaLgwtHY2U(grid, new EGallery.CreateHandler(grid_CellCreating));
		wsp7wfIE7ZRpcvBnnaiZ.qOys5bJ94(grid, right, wsp7wfIE7ZRpcvBnnaiZ.GIKIEAn16d7);
		right.Create = right_Create;
		int num = 2;
		while (true)
		{
			switch (num)
			{
			case 4:
				m12mW84KSKFWhFqvUAv.qOys5bJ94(PLogin, Login.Items, m12mW84KSKFWhFqvUAv.WDc4LgEQrE);
				BQfcMGnoPQ6NKh97YXO.qOys5bJ94(PLogin, doStat, BQfcMGnoPQ6NKh97YXO.PUQn6PIu1H);
				goto IL_06fa;
			case 3:
				Cb5RwHnlUG2JGJtoLqv.qOys5bJ94(N8OekY7ZSs0VxIjnUXP.qOys5bJ94(PLevel, N8OekY7ZSs0VxIjnUXP.Hyu7BoFvRk), new EItem("", A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x65271B85 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_8d18136f38d74e8f97d86f9cab9bb6ca)), Cb5RwHnlUG2JGJtoLqv.rManFS4CxJ);
				UDuUoC7QGXMEwiF2PmV.qOys5bJ94(PFlag, Enumer.Sale.Flag, UDuUoC7QGXMEwiF2PmV.YMi7kFiQMy);
				Cb5RwHnlUG2JGJtoLqv.qOys5bJ94(N8OekY7ZSs0VxIjnUXP.qOys5bJ94(PFlag, N8OekY7ZSs0VxIjnUXP.Hyu7BoFvRk), new EItem("", A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x381FC2BE ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9bf27d58a5584b2cb3d10730eccbd9c5)), Cb5RwHnlUG2JGJtoLqv.rManFS4CxJ);
				UDuUoC7QGXMEwiF2PmV.qOys5bJ94(PGrade, Enumer.Sale.Grade, UDuUoC7QGXMEwiF2PmV.YMi7kFiQMy);
				Cb5RwHnlUG2JGJtoLqv.qOys5bJ94(N8OekY7ZSs0VxIjnUXP.qOys5bJ94(PGrade, N8OekY7ZSs0VxIjnUXP.Hyu7BoFvRk), new EItem("", A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x17B67429 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_443e3241874d456eafc415ea3ca1dc23)), Cb5RwHnlUG2JGJtoLqv.rManFS4CxJ);
				UDuUoC7QGXMEwiF2PmV.qOys5bJ94(PType, new EItemCollection(new EItem((string)AR4Zasyr5NWWQ9bPNWpT(-1978347116 ^ -1924345438 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d77468cda0b945e3a4061d84efc82d59), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x177B8E3 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_dfe08b55d34746fe8ba0141c88660b08)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x41BDA654 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_141040ccb76546af8ef08c87257c6ac9), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x581F772F ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ed88ade44dea498194ae1c8221333918)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x26BE9451 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_e6be942033a74c0baffa094b080e9e6b), (string)AR4Zasyr5NWWQ9bPNWpT(0x75CD6CE ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d77468cda0b945e3a4061d84efc82d59)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2E426957 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_134d534497084f54beb0d98e0aef07e3), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x50849325 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_fca4b5827c154beda800a45f1a8c082c)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4CD12F21 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_282f1df2aaef433d8ead5f2d044ac539), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x526FBCE2 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_064ce8ea335043af8ca584aaced88385)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x5A1B8699 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9d8cead7b8414ddc80a1dd03ac024951), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x155CC5A7 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_24274e778cb34f0b9c6e5b5bbe7cc59b)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6914074F ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_88a9b435230c4eaf827b56a79d6ee53c), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4CD130C5 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_282f1df2aaef433d8ead5f2d044ac539))), UDuUoC7QGXMEwiF2PmV.YMi7kFiQMy);
				m12mW84KSKFWhFqvUAv.qOys5bJ94(PSize, new EItemCollection(new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x11FB223A ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9cd5a0aaeba04bb4b081b4b5fe90a190), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x36B672A5 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_12046703b6504e42927f2daf3fea0cdd)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7FBD0E47 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_603178892e9e413b98f3e6aa3da54b45), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6523F911 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_8d18136f38d74e8f97d86f9cab9bb6ca)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x68933208 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ce1dfb55258f4152afc31779feefbcba), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x75C2196 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d77468cda0b945e3a4061d84efc82d59)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1421C7C8 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_888e5b7a9cff40dea01c8d4cddb988ed), (string)AR4Zasyr5NWWQ9bPNWpT(0x6893325C ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ce1dfb55258f4152afc31779feefbcba))), m12mW84KSKFWhFqvUAv.WDc4LgEQrE);
				num = 1;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_de4c858d184a45e7a9ea91377a1f9cba != 0)
				{
					num = 1;
				}
				continue;
			default:
				XfIZVt2pHl1IJLDUwZx.qOys5bJ94(btn状态, 设置状态, XfIZVt2pHl1IJLDUwZx.FZQ21lPq5g);
				XfIZVt2pHl1IJLDUwZx.qOys5bJ94(btn分类, 批量分类, XfIZVt2pHl1IJLDUwZx.FZQ21lPq5g);
				if (PoloKgIlHfFSGOb26O2G.qOys5bJ94(Member.Limit.产品, PoloKgIlHfFSGOb26O2G.honIliyPVXI))
				{
					AIEQImtl6o6GfJuhUOe.qOys5bJ94(btn删除, true, AIEQImtl6o6GfJuhUOe.OVEtFTn5qP);
					XfIZVt2pHl1IJLDUwZx.qOys5bJ94(btn删除, 产品删除, XfIZVt2pHl1IJLDUwZx.FZQ21lPq5g);
				}
				m12mW84KSKFWhFqvUAv.qOys5bJ94(selOP, new EItemCollection(new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x328CA482 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_1599679ee1724b48bf7a40ba1cf1dff5), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x296E18E ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_21cb4653d1d542f1b694f9137faa29e1)), new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x40E729EB ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_943c194482ce4adf8f2a36eb84b85a8d), (string)AR4Zasyr5NWWQ9bPNWpT(0x119EA8C7 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0b5241575b984d088258edd0dec0807d), isFFjwkqk3pcZxCGKXY.qOys5bJ94(isFFjwkqk3pcZxCGKXY.moRkvPHrJR)), new EItem((string)AR4Zasyr5NWWQ9bPNWpT(0x328BD252 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_1599679ee1724b48bf7a40ba1cf1dff5), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x30135E0 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_1dfe6a356d324a4aba1c21d85b6856ea), MvDIhQhf8kkmvZ4Uerf.qOys5bJ94(MvDIhQhf8kkmvZ4Uerf.SwohCFdEsQ))), m12mW84KSKFWhFqvUAv.WDc4LgEQrE);
				m9oCu84lOv96NbjDuRB.qOys5bJ94(selOP, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x155BB301 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_24274e778cb34f0b9c6e5b5bbe7cc59b), m9oCu84lOv96NbjDuRB.Une4FfUblU);
				XfIZVt2pHl1IJLDUwZx.qOys5bJ94(btn导入主库, btn导入主库_Click, XfIZVt2pHl1IJLDUwZx.FZQ21lPq5g);
				kV7JyHujvdlDH8k0kuP.qOys5bJ94(btnGO, ETheme.Verdana, kV7JyHujvdlDH8k0kuP.lqnubJTQCZ);
				XfIZVt2pHl1IJLDUwZx.qOys5bJ94(btnGO, delegate
				{
					string text = cFiwTD6pIRgdHJh4tcE.qOys5bJ94(selOP, cFiwTD6pIRgdHJh4tcE.BMU61bIKPU);
					if (ccuqBrl774FBBcBTcHX.qOys5bJ94(text, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7C2B7EAD ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_8ae4a75652844f3fabca994c5d97cf24), ccuqBrl774FBBcBTcHX.lpSlAEh9YT))
					{
						qXuYODFTYbD2eHsJK31.qOys5bJ94(任务.Instance, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1425FB82 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_888e5b7a9cff40dea01c8d4cddb988ed), qXuYODFTYbD2eHsJK31.y15Fn7pQ5t);
					}
					else if (!ccuqBrl774FBBcBTcHX.qOys5bJ94(text, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2EE7494C ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0f5d594c7dcb4a7f9e6cc9d946db1eec), ccuqBrl774FBBcBTcHX.lpSlAEh9YT))
					{
						if (ccuqBrl774FBBcBTcHX.qOys5bJ94(text, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-1978347116 ^ -87451488 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_82991c292d16419fb0eb42340f9663db), ccuqBrl774FBBcBTcHX.lpSlAEh9YT))
						{
							qXuYODFTYbD2eHsJK31.qOys5bJ94(任务.Instance, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x185E1B95 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_a0f826422b7c442e91da261118a82e4c), qXuYODFTYbD2eHsJK31.y15Fn7pQ5t);
						}
					}
					else
					{
						qXuYODFTYbD2eHsJK31.qOys5bJ94(任务.Instance, (string)AR4Zasyr5NWWQ9bPNWpT(0x7581DB6 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d77468cda0b945e3a4061d84efc82d59), qXuYODFTYbD2eHsJK31.y15Fn7pQ5t);
					}
				}, XfIZVt2pHl1IJLDUwZx.FZQ21lPq5g);
				return;
			case 2:
				UDuUoC7QGXMEwiF2PmV.qOys5bJ94(PStat, Enumer.Sale.Stat, UDuUoC7QGXMEwiF2PmV.YMi7kFiQMy);
				Cb5RwHnlUG2JGJtoLqv.qOys5bJ94(N8OekY7ZSs0VxIjnUXP.qOys5bJ94(PStat, N8OekY7ZSs0VxIjnUXP.Hyu7BoFvRk), new EItem("", A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x119E763B ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0b5241575b984d088258edd0dec0807d)), Cb5RwHnlUG2JGJtoLqv.rManFS4CxJ);
				UDuUoC7QGXMEwiF2PmV.qOys5bJ94(PLevel, Enumer.Sale.Level, UDuUoC7QGXMEwiF2PmV.YMi7kFiQMy);
				num = 3;
				continue;
			case 1:
				m9oCu84lOv96NbjDuRB.qOys5bJ94(PSize, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x780337B9 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_488581fc99e442088402176de3de7cbd), m9oCu84lOv96NbjDuRB.Une4FfUblU);
				BQfcMGnoPQ6NKh97YXO.qOys5bJ94(PSize, doSelect, BQfcMGnoPQ6NKh97YXO.PUQn6PIu1H);
				zB1q6n7tdGDIkCKYMHP.qOys5bJ94(PStat, doFlag, zB1q6n7tdGDIkCKYMHP.yI87suyLyR);
				zB1q6n7tdGDIkCKYMHP.qOys5bJ94(PFlag, doLevel, zB1q6n7tdGDIkCKYMHP.yI87suyLyR);
				zB1q6n7tdGDIkCKYMHP.qOys5bJ94(PLevel, doGrade, zB1q6n7tdGDIkCKYMHP.yI87suyLyR);
				zB1q6n7tdGDIkCKYMHP.qOys5bJ94(PGrade, doType, zB1q6n7tdGDIkCKYMHP.yI87suyLyR);
				zB1q6n7tdGDIkCKYMHP.qOys5bJ94(PType, doSelect, zB1q6n7tdGDIkCKYMHP.yI87suyLyR);
				mYW7oA9etrxP4vNqBSC.qOys5bJ94(PBan, doStat, mYW7oA9etrxP4vNqBSC.bEd99i5xyZ);
				LyH2MYTSaIVO2r9y5OC.qOys5bJ94(PKind, KindLocal.采集分类, LyH2MYTSaIVO2r9y5OC.f8XTJtIGyl);
				Y9qcS150lHaWqSn5JJy.qOys5bJ94(PKind, doStat, Y9qcS150lHaWqSn5JJy.QJo5cTlCCX);
				PFek6qIPoZ26IOhHAI7R.qOys5bJ94(PDate, doStat, PFek6qIPoZ26IOhHAI7R.wkHIP605ZPA);
				if (TnOCCsZypUav3H1wRaH.qOys5bJ94(Member.Limit.产品, TnOCCsZypUav3H1wRaH.IJlZhh3TkJ) && !Member.只看自己产品)
				{
					AIEQImtl6o6GfJuhUOe.qOys5bJ94(PLogin, true, AIEQImtl6o6GfJuhUOe.OVEtFTn5qP);
					num = 4;
					if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_25089bab62a74b43b37fc67ad5fef7a9 != 0)
					{
						num = 4;
					}
					continue;
				}
				goto IL_06fa;
			case 5:
				{
					BQfcMGnoPQ6NKh97YXO.qOys5bJ94(PArea, doSelect, BQfcMGnoPQ6NKh97YXO.PUQn6PIu1H);
					goto IL_054b;
				}
				IL_054b:
				Pw9twT4t5esjrJJuAun.qOys5bJ94(PKey, doStat, Pw9twT4t5esjrJJuAun.svP4sm6v5c);
				Pw9twT4t5esjrJJuAun.qOys5bJ94(PWord, doStat, Pw9twT4t5esjrJJuAun.svP4sm6v5c);
				PVK4D4CatCgKCjvOKdH.qOys5bJ94(PWord, doVoid, PVK4D4CatCgKCjvOKdH.hXLCNHe4M6);
				mYW7oA9etrxP4vNqBSC.qOys5bJ94(chkSelect, delegate
				{
					bool flag = PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(chkSelect, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0);
					Qw9jC1ITHWsCwqvA3oNj.qOys5bJ94(grid, flag, Qw9jC1ITHWsCwqvA3oNj.er2ITiRaVXc);
					pxHVUshl48ahyalcwFt.qOys5bJ94(lblAll, flag ? CuJwqoitNBAPSK8LlTo.qOys5bJ94(LNSkANLDPq4fpf6PlDP.qOys5bJ94(LNSkANLDPq4fpf6PlDP.YjZLRD46tm), CuJwqoitNBAPSK8LlTo.UAnisu1bsS) : mBCqVxIrIfluJgiV1u1f.qOys5bJ94(LNSkANLDPq4fpf6PlDP.qOys5bJ94(LNSkANLDPq4fpf6PlDP.YjZLRD46tm), mBCqVxIrIfluJgiV1u1f.OanIr2ladRm), pxHVUshl48ahyalcwFt.pgchF9iJkR);
					if (!flag)
					{
						jLTKnh2z5j5GYdWFTT5.qOys5bJ94(lblAll, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2E43D235 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_134d534497084f54beb0d98e0aef07e3), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
					}
				}, mYW7oA9etrxP4vNqBSC.bEd99i5xyZ);
				pxHVUshl48ahyalcwFt.qOys5bJ94(lblAll, mBCqVxIrIfluJgiV1u1f.qOys5bJ94(LNSkANLDPq4fpf6PlDP.qOys5bJ94(LNSkANLDPq4fpf6PlDP.YjZLRD46tm), mBCqVxIrIfluJgiV1u1f.OanIr2ladRm), pxHVUshl48ahyalcwFt.pgchF9iJkR);
				MqjAkSvaUQjSlBnQkQ1.qOys5bJ94(lblAll, delegate
				{
					if (!PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(chkSelect, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0))
					{
						EGallery eGallery = grid;
						bool flag;
						q2jfXx5rEKUh24rCAxJ.qOys5bJ94(chkSelect, flag = true, q2jfXx5rEKUh24rCAxJ.uRO5wuVmVy);
						Qw9jC1ITHWsCwqvA3oNj.qOys5bJ94(eGallery, flag, Qw9jC1ITHWsCwqvA3oNj.er2ITiRaVXc);
					}
					bool flag2 = ccuqBrl774FBBcBTcHX.qOys5bJ94(y23J1H2DqWXElNIxLR5.qOys5bJ94(lblAll, y23J1H2DqWXElNIxLR5.nsm2RZKktA), (string)AR4Zasyr5NWWQ9bPNWpT(0x328F9ED0 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_1599679ee1724b48bf7a40ba1cf1dff5), ccuqBrl774FBBcBTcHX.lpSlAEh9YT);
					JP225uITXGCLS4MnQdnJ.qOys5bJ94(grid, flag2, JP225uITXGCLS4MnQdnJ.VrWITgE2kfr);
					jLTKnh2z5j5GYdWFTT5.qOys5bJ94(lblAll, flag2 ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4850E103 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9e7c8ea59e0745258e93689531c631f8) : A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3D9E6B27 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b912e71988d0440f99d561d4df651c0e), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				}, MqjAkSvaUQjSlBnQkQ1.eDAvN0mSIY);
				PeKMKRcXlHaGFEAEj9v.qOys5bJ94(PTurn, doFresh, PeKMKRcXlHaGFEAEj9v.NZhcg2QfMP);
				XfIZVt2pHl1IJLDUwZx.qOys5bJ94(btn多选导入, Btn多选导入_Click, XfIZVt2pHl1IJLDUwZx.FZQ21lPq5g);
				num = 0;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_c40cd20adffa4674bd5e67607539d20c != 0)
				{
					num = 0;
				}
				continue;
				IL_06fa:
				G7uvoHI548Ujp814hqyR.qOys5bJ94(PSource, PSource_GetMiddles, G7uvoHI548Ujp814hqyR.amrI5mYHFSH);
				G2bjsLI5QmblZgOoN6Md.qOys5bJ94(PSource, PSource_GetKinder, G2bjsLI5QmblZgOoN6Md.e2tI5kyEYXs);
				RinqEAI5ZByn2FYaXxFt.qOys5bJ94(PSource, Enumer.Web.Items, RinqEAI5ZByn2FYaXxFt.NQxI5BTKDQ4);
				s5XlbiI5qxc1w9rbRxbu.qOys5bJ94(PSource, doSelect, s5XlbiI5qxc1w9rbRxbu.t7FI5vYgZKg);
				if (dEBJd9ZKagHn6TtOl0L.qOys5bJ94(dEBJd9ZKagHn6TtOl0L.O6OZLYjOCM) || fGowtDQ4wg8vJOUxlX8.qOys5bJ94(fGowtDQ4wg8vJOUxlX8.BuvQmwv7hq))
				{
					break;
				}
				goto IL_054b;
			}
			AIEQImtl6o6GfJuhUOe.qOys5bJ94(PArea, true, AIEQImtl6o6GfJuhUOe.OVEtFTn5qP);
			m12mW84KSKFWhFqvUAv.qOys5bJ94(PArea, Enumer.Meli.Abbrs, m12mW84KSKFWhFqvUAv.WDc4LgEQrE);
			int num2 = 5;
			num = num2;
		}
	}

	private EItemCollection PSource_GetMiddles(EItem it)
	{
		return fAOyoNUHMY49x2k29oc.qOys5bJ94(it, fAOyoNUHMY49x2k29oc.rwIUiByXs8) switch
		{
			1 => K4HnFrptNMs9w0jbZf3.qOys5bJ94(Amazon.Areas, false, K4HnFrptNMs9w0jbZf3.KZDpsRiuvb), 
			8 => K4HnFrptNMs9w0jbZf3.qOys5bJ94(Enumer.Meli.Abbrs, false, K4HnFrptNMs9w0jbZf3.KZDpsRiuvb), 
			_ => null, 
		};
	}

	private IKind PSource_GetKinder(EItem it, EItem mid)
	{
		switch (fAOyoNUHMY49x2k29oc.qOys5bJ94(it, fAOyoNUHMY49x2k29oc.rwIUiByXs8))
		{
		case 1:
		{
			KindAmazon kindAmazon = mdh9vMXeQPSZKE9v8gR.qOys5bJ94(mlHdPvyQLH3c60pXHww.qOys5bJ94(mid, mlHdPvyQLH3c60pXHww.xiPykdKX24), null, mdh9vMXeQPSZKE9v8gR.tk6X9cU6nV);
			if (kindAmazon == null)
			{
				return null;
			}
			return d7gGfQI5tF5fwTbR1l7m.qOys5bJ94(kindAmazon, d7gGfQI5tF5fwTbR1l7m.tDcI5sCKCAH);
		}
		case 2:
		{
			KindAliExpress instance = KindAliExpress.Instance;
			if (instance == null)
			{
				int num = 0;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_25089bab62a74b43b37fc67ad5fef7a9 != 0)
				{
					num = 0;
				}
				return num switch
				{
					_ => (IKind)null, 
				};
			}
			return d7gGfQI5tF5fwTbR1l7m.qOys5bJ94(instance, d7gGfQI5tF5fwTbR1l7m.tDcI5sCKCAH);
		}
		case 12:
		{
			KindShopee kindShopee = fNt3exUaHXUxZMKEQi8.qOys5bJ94(fNt3exUaHXUxZMKEQi8.E1uUN45gv5);
			if (kindShopee == null)
			{
				return null;
			}
			return d7gGfQI5tF5fwTbR1l7m.qOys5bJ94(kindShopee, d7gGfQI5tF5fwTbR1l7m.tDcI5sCKCAH);
		}
		case 8:
		{
			KindMeli kindMeli = E8PlMsUtJVqiN7SuGvT.qOys5bJ94(mlHdPvyQLH3c60pXHww.qOys5bJ94(mid, mlHdPvyQLH3c60pXHww.xiPykdKX24), E8PlMsUtJVqiN7SuGvT.sCIUsdJ7OB);
			if (kindMeli == null)
			{
				return null;
			}
			return d7gGfQI5tF5fwTbR1l7m.qOys5bJ94(kindMeli, d7gGfQI5tF5fwTbR1l7m.tDcI5sCKCAH);
		}
		default:
			return null;
		}
	}

	public override void Load()
	{
		doTotal(this, EventArgs.Empty);
	}

	private ICell grid_CellCreating(EGallery g)
	{
		return new EGSale(g);
	}

	private EControl right_Create(ISource src, string name)
	{
		return new ZYing.Detail.Sale(src, name, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4857FC75 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9e7c8ea59e0745258e93689531c631f8));
	}

	private void whole_SizeChanged(object sender, EventArgs e)
	{
		if (I69Zs8IESLu15SmrOZ62.qOys5bJ94(right, I69Zs8IESLu15SmrOZ62.LyFIEJHrSSD) == 0 || hKTUAtNetEpBhxVuqmA.qOys5bJ94(this, hKTUAtNetEpBhxVuqmA.vqON9H9w4p) == 0)
		{
			return;
		}
		int num = hKTUAtNetEpBhxVuqmA.qOys5bJ94(this, hKTUAtNetEpBhxVuqmA.vqON9H9w4p) * I69Zs8IESLu15SmrOZ62.qOys5bJ94(right, I69Zs8IESLu15SmrOZ62.LyFIEJHrSSD) / 100;
		int num2 = hKTUAtNetEpBhxVuqmA.qOys5bJ94(this, hKTUAtNetEpBhxVuqmA.vqON9H9w4p) - CqZXDwIEzUHqGYJiYBvP.qOys5bJ94(right, CqZXDwIEzUHqGYJiYBvP.tddIjOsgVNU);
		if (num < num2)
		{
			num = num2;
		}
		if (payrf4wMtgsKaUBXUbS.qOys5bJ94(right, payrf4wMtgsKaUBXUbS.hFcwxSFOmn) != num)
		{
			int num3 = 0;
			if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_7fe63cd1d69a404ebc71b22a4b9c5897 != 0)
			{
				num3 = 0;
			}
			switch (num3)
			{
			}
			Nvk9N3nWojUm526Uph9.qOys5bJ94(right, num, Nvk9N3nWojUm526Uph9.O6TnEW8gJC);
			Wav4YXIFzkbHfXl4wvGt.qOys5bJ94(right, payrf4wMtgsKaUBXUbS.qOys5bJ94(right, payrf4wMtgsKaUBXUbS.hFcwxSFOmn) * 100 / hKTUAtNetEpBhxVuqmA.qOys5bJ94(this, hKTUAtNetEpBhxVuqmA.vqON9H9w4p), Wav4YXIFzkbHfXl4wvGt.YSkI4ONZBrJ);
		}
	}

	private Para doPara()
	{
		Para para = new Para
		{
			ban = PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(PBan, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0),
			localid = WYn8YtIIfdYZQxBWsS1i.qOys5bJ94(PKind, WYn8YtIIfdYZQxBWsS1i.uPlIICSKiNc),
			start = cFiwTD6pIRgdHJh4tcE.qOys5bJ94(ymF6QQ3pXS3WmDL1oNg.qOys5bJ94(PDate, ymF6QQ3pXS3WmDL1oNg.WQA31owjVW), cFiwTD6pIRgdHJh4tcE.BMU61bIKPU),
			end = cFiwTD6pIRgdHJh4tcE.qOys5bJ94(ATteTW3Dj9TRuHaIWmy.qOys5bJ94(PDate, ATteTW3Dj9TRuHaIWmy.tiv3R1Ays1), cFiwTD6pIRgdHJh4tcE.BMU61bIKPU),
			loginid = lGeGLwLMWcI3YHQqfYq.qOys5bJ94(cFiwTD6pIRgdHJh4tcE.qOys5bJ94(PLogin, cFiwTD6pIRgdHJh4tcE.BMU61bIKPU), lGeGLwLMWcI3YHQqfYq.VOnLx3gIY8),
			from = lGeGLwLMWcI3YHQqfYq.qOys5bJ94(LlqbPMI5WbkWUDGm2IOA.qOys5bJ94(PSource, LlqbPMI5WbkWUDGm2IOA.RZNI5EUN32X), lGeGLwLMWcI3YHQqfYq.VOnLx3gIY8),
			extra = oH7lYJI5juyOClHoIsD0.qOys5bJ94(PSource, oH7lYJI5juyOClHoIsD0.jgSI5bLjC0a),
			area = CYyUfWI5orteuKExOkgG.qOys5bJ94(PSource, CYyUfWI5orteuKExOkgG.PJkI56DKCmZ),
			low = gQ63UAMZa9j0mdcdR4a.qOys5bJ94(NnQqi01aW8l9hlBTBRI.qOys5bJ94(Price, NnQqi01aW8l9hlBTBRI.wtn1N1jDRH), gQ63UAMZa9j0mdcdR4a.QI0MBRFToI),
			high = gQ63UAMZa9j0mdcdR4a.qOys5bJ94(Ff8I3D1P3luUBrmG1by.qOys5bJ94(Price, Ff8I3D1P3luUBrmG1by.YXA1ux55tY), gQ63UAMZa9j0mdcdR4a.QI0MBRFToI),
			key = KD3OclIsSMmJbPHCBZY6.qOys5bJ94(y23J1H2DqWXElNIxLR5.qOys5bJ94(PKey, y23J1H2DqWXElNIxLR5.nsm2RZKktA), '，', ',', KD3OclIsSMmJbPHCBZY6.cL7IsJ2Dy99),
			word = y23J1H2DqWXElNIxLR5.qOys5bJ94(PWord, y23J1H2DqWXElNIxLR5.nsm2RZKktA),
			stat = lGeGLwLMWcI3YHQqfYq.qOys5bJ94(vXrOiWIbaSTvtNliL0Uf.qOys5bJ94(PStat, vXrOiWIbaSTvtNliL0Uf.EPmIbN4NmE8), lGeGLwLMWcI3YHQqfYq.VOnLx3gIY8),
			level = lGeGLwLMWcI3YHQqfYq.qOys5bJ94(vXrOiWIbaSTvtNliL0Uf.qOys5bJ94(PLevel, vXrOiWIbaSTvtNliL0Uf.EPmIbN4NmE8), lGeGLwLMWcI3YHQqfYq.VOnLx3gIY8),
			flag = lGeGLwLMWcI3YHQqfYq.qOys5bJ94(vXrOiWIbaSTvtNliL0Uf.qOys5bJ94(PFlag, vXrOiWIbaSTvtNliL0Uf.EPmIbN4NmE8), lGeGLwLMWcI3YHQqfYq.VOnLx3gIY8),
			grade = lGeGLwLMWcI3YHQqfYq.qOys5bJ94(vXrOiWIbaSTvtNliL0Uf.qOys5bJ94(PGrade, vXrOiWIbaSTvtNliL0Uf.EPmIbN4NmE8), lGeGLwLMWcI3YHQqfYq.VOnLx3gIY8),
			type = lGeGLwLMWcI3YHQqfYq.qOys5bJ94(vXrOiWIbaSTvtNliL0Uf.qOys5bJ94(PType, vXrOiWIbaSTvtNliL0Uf.EPmIbN4NmE8), lGeGLwLMWcI3YHQqfYq.VOnLx3gIY8),
			pagesize = lGeGLwLMWcI3YHQqfYq.qOys5bJ94(cFiwTD6pIRgdHJh4tcE.qOys5bJ94(PSize, cFiwTD6pIRgdHJh4tcE.BMU61bIKPU), lGeGLwLMWcI3YHQqfYq.VOnLx3gIY8),
			values = th2rvQI5rgjmxUMhnxRI.qOys5bJ94(PSource, th2rvQI5rgjmxUMhnxRI.yOhI5wOSvjI)
		};
		if (para.values != null)
		{
			string text = CYyUfWI5orteuKExOkgG.qOys5bJ94(PSource, CYyUfWI5orteuKExOkgG.PJkI56DKCmZ);
			for (int i = 0; i < para.values.Length; i++)
			{
				para.values[i] = GkSwMgtX7pQMRBHMc6S.qOys5bJ94(aGThQ0zBsC1uu6CBnB.qOys5bJ94(ref para.from, aGThQ0zBsC1uu6CBnB.lvfIO2Ntg8), (ja0HD4afsGXTlOU1r5g.qOys5bJ94(text, ja0HD4afsGXTlOU1r5g.sBEaCZVGXB) > 0) ? l6GFAG7hp5Ui898HhZ.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1425BC06 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_888e5b7a9cff40dea01c8d4cddb988ed), text, l6GFAG7hp5Ui898HhZ.iIvA7JOfD) : "", A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2EE70EAC ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0f5d594c7dcb4a7f9e6cc9d946db1eec), para.values[i], GkSwMgtX7pQMRBHMc6S.h5qtgPYcVJ);
			}
		}
		return para;
	}

	private void doVoid(object sender, EventArgs e)
	{
		m9oCu84lOv96NbjDuRB.qOys5bJ94(PKind, "", m9oCu84lOv96NbjDuRB.Une4FfUblU);
		VZajbfezP2MhcsK1XbD.qOys5bJ94(PDate, VZajbfezP2MhcsK1XbD.snB9OP9dPv);
		VZajbfezP2MhcsK1XbD.qOys5bJ94(PLogin, VZajbfezP2MhcsK1XbD.snB9OP9dPv);
		jLTKnh2z5j5GYdWFTT5.qOys5bJ94(PKey, "", jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
		jLTKnh2z5j5GYdWFTT5.qOys5bJ94(PWord, "", jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
		EFilter pStat = PStat;
		EFilter pFlag = PFlag;
		string text;
		XaB2LAIbPe7yOZPRyZLR.qOys5bJ94(PGrade, text = "", XaB2LAIbPe7yOZPRyZLR.aeAIbuCrjhE);
		string text2;
		XaB2LAIbPe7yOZPRyZLR.qOys5bJ94(pFlag, text2 = text, XaB2LAIbPe7yOZPRyZLR.aeAIbuCrjhE);
		XaB2LAIbPe7yOZPRyZLR.qOys5bJ94(pStat, text2, XaB2LAIbPe7yOZPRyZLR.aeAIbuCrjhE);
		XaB2LAIbPe7yOZPRyZLR.qOys5bJ94(PType, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x30342353 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_7a1fed405f4748f2b7b141bf9600fe2e), XaB2LAIbPe7yOZPRyZLR.aeAIbuCrjhE);
		doTotal(sender, e);
	}

	private void doTotal(object sender, EventArgs e)
	{
		doCommand(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7C909992 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_c5ba8cdf0efd483d97721d48bfc441f0), doPara());
	}

	private void doStat(object sender, EventArgs e)
	{
		doCommand(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7F5EDBB8 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_71b15132f4c3474aaebfe1530c68df31), doPara());
	}

	private void doLevel(object sender, EventArgs e)
	{
		doCommand((string)AR4Zasyr5NWWQ9bPNWpT(0x4F880B06 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b480309f51bd48c194e1ed3315e9d650), doPara());
	}

	private void doFlag(object sender, EventArgs e)
	{
		doCommand(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0xFDA4B69 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_bf79aafc640a4fee8effe87c86aa1398), doPara());
	}

	private void doGrade(object sender, EventArgs e)
	{
		doCommand(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3A45796E ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_c0fff80a6c6a4ab1825b648985b91c3a), doPara());
	}

	private void doType(object sender, EventArgs e)
	{
		doCommand(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x119EE4E1 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0b5241575b984d088258edd0dec0807d), doPara());
	}

	private void doSelect(object sender, EventArgs e)
	{
		doCommand(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3D9DA355 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b912e71988d0440f99d561d4df651c0e), doPara());
	}

	private void doFresh(object sender, EventArgs e)
	{
		Para para = doPara();
		para.maxcount = wlBbfQ7KjpA0R9ZNBP4.qOys5bJ94(PTurn, wlBbfQ7KjpA0R9ZNBP4.aVB7LkDB7C);
		para.page = jij2HW7lW8W42D8YBfB.qOys5bJ94(PTurn, jij2HW7lW8W42D8YBfB.POs7F11maw);
		doCommand((string)AR4Zasyr5NWWQ9bPNWpT(0x3D9437C8 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ec48559fcf5644799b4f514a052ee03b), para);
	}

	[AsyncStateMachine(typeof(<doCommand>d__22))]
	private void doCommand(string cmd, Para p)
	{
		<doCommand>d__22 stateMachine = default;
		stateMachine.<>t__builder = NoCSc1ySh7ioNVnK3b8.qOys5bJ94(NoCSc1ySh7ioNVnK3b8.EuOyJiAorm);
		stateMachine.<>4__this = this;
		stateMachine.cmd = cmd;
		stateMachine.p = p;
		stateMachine.<>1__state = -1;
		stateMachine.<>t__builder.Start(ref stateMachine);
	}

	private void 设置状态(object sender, EventArgs e)
	{
		if (!PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(chkSelect, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0))
		{
			raZqQm15icvksyfKy1g.qOys5bJ94(chkSelect, true, raZqQm15icvksyfKy1g.hjx1UOoNFG);
			return;
		}
		int num = 0;
		for (int i = 0; i < jtwfoCIToQl9n5ohwkvu.qOys5bJ94(grid, jtwfoCIToQl9n5ohwkvu.CdqIT6mbZyE); i++)
		{
			if ((dPegXZITrT0v8bhnlqHy.qOys5bJ94(grid, i, dPegXZITrT0v8bhnlqHy.C3BITwTmsVc) as EGSale).Selected)
			{
				num++;
			}
		}
		if (num <= 0)
		{
			int num2 = 0;
			if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d59bc24a307e4bdca17ad3010cacc7b5 != 0)
			{
				num2 = 0;
			}
			switch (num2)
			{
			}
			PWGKTQ28i8QujJOnnX4.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x336319C8 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9eb1b575102d43aeb17e7aaeef3dc5dd), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x74E93408 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d59bc24a307e4bdca17ad3010cacc7b5), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
		}
		else
		{
			产品状态Dialog instance = 产品状态Dialog.Instance;
			instance.Finish = (string op) =>
			{
				Batch状态(op);
			};
			instance.Note = MaKBHa0rb8pCci0Q1b.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x691480E9 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_88a9b435230c4eaf827b56a79d6ee53c), aGThQ0zBsC1uu6CBnB.qOys5bJ94(ref num, aGThQ0zBsC1uu6CBnB.lvfIO2Ntg8), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x40E48535 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_943c194482ce4adf8f2a36eb84b85a8d), MaKBHa0rb8pCci0Q1b.UNjcxHPJw);
			NLguUnXfTc6L2XcpPpx.qOys5bJ94(instance, NLguUnXfTc6L2XcpPpx.rZsXCAG9e1);
			instance.Show(this);
		}
	}

	[AsyncStateMachine(typeof(<产品合并>d__24))]
	private void 产品合并(object sender, EventArgs e)
	{
		<产品合并>d__24 stateMachine = default;
		stateMachine.<>t__builder = NoCSc1ySh7ioNVnK3b8.qOys5bJ94(NoCSc1ySh7ioNVnK3b8.EuOyJiAorm);
		stateMachine.<>4__this = this;
		stateMachine.<>1__state = -1;
		stateMachine.<>t__builder.Start(ref stateMachine);
	}

	[AsyncStateMachine(typeof(<Btn多选导入_Click>d__25))]
	private void Btn多选导入_Click(object sender, EventArgs e)
	{
		<Btn多选导入_Click>d__25 stateMachine = default;
		stateMachine.<>t__builder = NoCSc1ySh7ioNVnK3b8.qOys5bJ94(NoCSc1ySh7ioNVnK3b8.EuOyJiAorm);
		stateMachine.<>4__this = this;
		stateMachine.<>1__state = -1;
		stateMachine.<>t__builder.Start(ref stateMachine);
	}

	private void 挂号模板(object sender, EventArgs e)
	{
		if (PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(chkSelect, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0))
		{
			int num = 0;
			StringBuilder stringBuilder = new StringBuilder();
			for (int i = 0; i < jtwfoCIToQl9n5ohwkvu.qOys5bJ94(grid, jtwfoCIToQl9n5ohwkvu.CdqIT6mbZyE); i++)
			{
				EGSale eGSale = dPegXZITrT0v8bhnlqHy.qOys5bJ94(grid, i, dPegXZITrT0v8bhnlqHy.C3BITwTmsVc) as EGSale;
				if (eGSale.Selected)
				{
					yuX4loaHrxU06DUKTed.qOys5bJ94(stringBuilder, l6GFAG7hp5Ui898HhZ.qOys5bJ94((Cn1oVaaeOcrG1OdN6mm.qOys5bJ94(stringBuilder, Cn1oVaaeOcrG1OdN6mm.MVPa9bItHD) > 0) ? A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x17B65863 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_443e3241874d456eafc415ea3ca1dc23) : "", eGSale.Name, l6GFAG7hp5Ui898HhZ.iIvA7JOfD), yuX4loaHrxU06DUKTed.KVFaiRKWhU);
					num++;
				}
			}
			if (num > 0)
			{
				运费设置 instance = 运费设置.Instance;
				instance.Finish = Batch挂号费;
				instance.Show(this, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7580036 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d77468cda0b945e3a4061d84efc82d59), null, rXYd8UaVChLCL8Lkm5d.qOys5bJ94(stringBuilder, rXYd8UaVChLCL8Lkm5d.T7ZadVdu8C));
				object obj = AR4Zasyr5NWWQ9bPNWpT(0x4F84FF4C ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_845968e715db46549494fad4e2bbf631);
				int num2 = IXieGTIny8wr2spEccVM.qOys5bJ94(grid, IXieGTIny8wr2spEccVM.Js8InhkLawi);
				instance.Note = MaKBHa0rb8pCci0Q1b.qOys5bJ94((string)obj, aGThQ0zBsC1uu6CBnB.qOys5bJ94(ref num2, aGThQ0zBsC1uu6CBnB.lvfIO2Ntg8), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2919604 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_21cb4653d1d542f1b694f9137faa29e1), MaKBHa0rb8pCci0Q1b.UNjcxHPJw);
			}
			else
			{
				PWGKTQ28i8QujJOnnX4.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2CB61B95 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0efbc81bc23c42efac0295bf66972dea), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x17729E3 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_dfe08b55d34746fe8ba0141c88660b08), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
			}
		}
		else
		{
			raZqQm15icvksyfKy1g.qOys5bJ94(chkSelect, true, raZqQm15icvksyfKy1g.hjx1UOoNFG);
		}
	}

	private void 批量分类(object sender, EventArgs e)
	{
		if (!PjRUQW9fyZdN4kNsG0Y.qOys5bJ94(chkSelect, PjRUQW9fyZdN4kNsG0Y.X899C85Ts0))
		{
			raZqQm15icvksyfKy1g.qOys5bJ94(chkSelect, true, raZqQm15icvksyfKy1g.hjx1UOoNFG);
			return;
		}
		int num = 0;
		for (int i = 0; i < jtwfoCIToQl9n5ohwkvu.qOys5bJ94(grid, jtwfoCIToQl9n5ohwkvu.CdqIT6mbZyE); i++)
		{
			if ((dPegXZITrT0v8bhnlqHy.qOys5bJ94(grid, i, dPegXZITrT0v8bhnlqHy.C3BITwTmsVc) as EGSale).Selected)
			{
				num++;
			}
		}
		int num2 = 0;
		if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0c2bf320e2b849e1b3895cf28bf5cedd == 0)
		{
			num2 = 0;
		}
		switch (num2)
		{
		}
		if (num <= 0)
		{
			PWGKTQ28i8QujJOnnX4.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x291965A ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_21cb4653d1d542f1b694f9137faa29e1), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x17729E3 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_dfe08b55d34746fe8ba0141c88660b08), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
			return;
		}
		D批量产品分类 instance = D批量产品分类.Instance;
		instance.Submit = Batch分类;
		instance.Note = MaKBHa0rb8pCci0Q1b.qOys5bJ94(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x30420A5D ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b6a459225bf544d6b93d37e0957a79c2), aGThQ0zBsC1uu6CBnB.qOys5bJ94(ref num, aGThQ0zBsC1uu6CBnB.lvfIO2Ntg8), (string)AR4Zasyr5NWWQ9bPNWpT(0x78038049 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_488581fc99e442088402176de3de7cbd), MaKBHa0rb8pCci0Q1b.UNjcxHPJw);
		NLguUnXfTc6L2XcpPpx.qOys5bJ94(instance, NLguUnXfTc6L2XcpPpx.rZsXCAG9e1);
		s10sJs3MeIBnud5lYn3.qOys5bJ94(instance, this, s10sJs3MeIBnud5lYn3.b753xiyhgW);
	}

	[AsyncStateMachine(typeof(<产品删除>d__28))]
	private void 产品删除(object sender, EventArgs e)
	{
		<产品删除>d__28 stateMachine = default;
		stateMachine.<>t__builder = NoCSc1ySh7ioNVnK3b8.qOys5bJ94(NoCSc1ySh7ioNVnK3b8.EuOyJiAorm);
		stateMachine.<>4__this = this;
		stateMachine.<>1__state = -1;
		stateMachine.<>t__builder.Start(ref stateMachine);
	}

	private void Batch挂号费(int shipid, NFreightCollection sc)
	{
		raZqQm15icvksyfKy1g.qOys5bJ94(chkSelect, false, raZqQm15icvksyfKy1g.hjx1UOoNFG);
		PWGKTQ28i8QujJOnnX4.qOys5bJ94((string)AR4Zasyr5NWWQ9bPNWpT(0x75CD4C2 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d77468cda0b945e3a4061d84efc82d59), A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x30347139 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_7a1fed405f4748f2b7b141bf9600fe2e), MessageBoxButtons.OK, MessageBoxIcon.Asterisk, PWGKTQ28i8QujJOnnX4.aya2YC5url);
	}

	[AsyncStateMachine(typeof(<Batch状态>d__30))]
	public void Batch状态(string op)
	{
		<Batch状态>d__30 stateMachine = default;
		stateMachine.<>t__builder = NoCSc1ySh7ioNVnK3b8.qOys5bJ94(NoCSc1ySh7ioNVnK3b8.EuOyJiAorm);
		stateMachine.<>4__this = this;
		stateMachine.op = op;
		stateMachine.<>1__state = -1;
		stateMachine.<>t__builder.Start(ref stateMachine);
	}

	[AsyncStateMachine(typeof(<Batch分类>d__31))]
	public void Batch分类(int localid)
	{
		<Batch分类>d__31 stateMachine = default;
		stateMachine.<>t__builder = NoCSc1ySh7ioNVnK3b8.qOys5bJ94(NoCSc1ySh7ioNVnK3b8.EuOyJiAorm);
		stateMachine.<>4__this = this;
		stateMachine.localid = localid;
		stateMachine.<>1__state = -1;
		stateMachine.<>t__builder.Start(ref stateMachine);
	}

	[AsyncStateMachine(typeof(<btn导入主库_Click>d__32))]
	private void btn导入主库_Click(object sender, EventArgs e)
	{
		<btn导入主库_Click>d__32 stateMachine = default;
		stateMachine.<>t__builder = NoCSc1ySh7ioNVnK3b8.qOys5bJ94(NoCSc1ySh7ioNVnK3b8.EuOyJiAorm);
		stateMachine.<>4__this = this;
		stateMachine.<>1__state = -1;
		stateMachine.<>t__builder.Start(ref stateMachine);
	}

	public ZYing.Detail.Sale CreateNew()
	{
		E2lXsDInKsiji9iJoZ77.qOys5bJ94(right, grid, A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-1978347116 ^ -665445058 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_10ffe2be534f4959b052b27cc6e82b39), E2lXsDInKsiji9iJoZ77.byPInLOIgxU);
		return IaHxJmIwaAOpTG8Pj91d.qOys5bJ94(right, IaHxJmIwaAOpTG8Pj91d.i6ZIwNhSPdB) as ZYing.Detail.Sale;
	}

	protected override void Dispose(bool disposing)
	{
		if (disposing && components != null)
		{
			B21j5m2afmkjUADHXaO.qOys5bJ94(components, B21j5m2afmkjUADHXaO.nuq2N0pFdF);
		}
		vkwOOpqHfB6EONehjSj.qOys5bJ94(this, disposing, vkwOOpqHfB6EONehjSj.LIDqi1oSWu);
	}

	private void InitializeComponent()
	{
		this.components = new System.ComponentModel.Container();
		this.right = new ZYing.UI.EPanelEx();
		this.main = new ZYing.UI.ELayer();
		this.PKey = new ZYing.UI.EText();
		this.PKind = new ZYing.UI.ELocalSearch();
		this.PLogin = new ZYing.UI.ECombo();
		this.Price = new ZYing.UI.ETextPair();
		this.txtDelays = new ZYing.UI.ETextPair();
		this.PDate = new ZYing.UI.ESelectDateRange();
		this.PArea = new ZYing.UI.ESelect();
		this.PSource = new ZYing.UI.EDropRootMidLinkageSearch();
		this.PSize = new ZYing.UI.ESelect();
		this.PWord = new ZYing.UI.ESearch();
		this.eLine1 = new ZYing.UI.ELine();
		this.PStat = new ZYing.UI.EFilter();
		this.PBan = new ZYing.UI.ECheck();
		this.PFlag = new ZYing.UI.EFilter();
		this.PLevel = new ZYing.UI.EFilter();
		this.PGrade = new ZYing.UI.EFilter();
		this.PType = new ZYing.UI.EFilter();
		this.grid = new ZYing.UI.EGallery();
		int num = 7;
		while (true)
		{
			int num2;
			switch (num)
			{
			case 17:
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PTurn, 8, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.PTurn, (string)ZYing.Channel.采集列表.AR4Zasyr5NWWQ9bPNWpT(0x2CB26EAF ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0efbc81bc23c42efac0295bf66972dea), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.lblTotal, ZYing.UI.FlowAlignment.Right, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				pxHVUshl48ahyalcwFt.qOys5bJ94(this.lblTotal, aQHLxQhKt15RVtJiKfo.qOys5bJ94(aQHLxQhKt15RVtJiKfo.PAchLoDpQ8), pxHVUshl48ahyalcwFt.pgchF9iJkR);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.lblTotal, new System.Drawing.Point(704, 524), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				num = 34;
				break;
			case 37:
				Q6lPi66j62cRc7qDDbh.qOys5bJ94(this.PArea, false, Q6lPi66j62cRc7qDDbh.GHf6b5rc0v);
				wH28it4fIBHr55P7oYo.qOys5bJ94(this.PSource, true, wH28it4fIBHr55P7oYo.fBL4C8jHGd);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.PSource, ZYing.UI.FlowAlignment.Justify, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PSource, new System.Drawing.Point(11, 77), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				jxjTTThDV9tTypDLsCq.qOys5bJ94(this.PSource, new System.Drawing.Size(432, 27), jxjTTThDV9tTypDLsCq.fPshRMocUh);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PSource, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4F8DD5F6 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b480309f51bd48c194e1ed3315e9d650), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PSource, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				KJTuOrI5M9rM8ZZsjLUp.qOys5bJ94(this.PSource, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2EE38010 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0f5d594c7dcb4a7f9e6cc9d946db1eec), KJTuOrI5M9rM8ZZsjLUp.ariI5x0xsin);
				SwPF3sI5TkCZnQsmaEwr.qOys5bJ94(this.PSource, 140, SwPF3sI5TkCZnQsmaEwr.BYTI5nMxd4F);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PSource, new System.Drawing.Size(838, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				yEIkvEI55RKTNjbV9iBR.qOys5bJ94(this.PSource, 693, yEIkvEI55RKTNjbV9iBR.xiQI5UUDU0R);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PSource, 244, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				wH28it4fIBHr55P7oYo.qOys5bJ94(this.PSize, true, wH28it4fIBHr55P7oYo.fBL4C8jHGd);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PSize, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PSize, new System.Drawing.Point(11, 111), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PSize, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x70DA6B82 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_82991c292d16419fb0eb42340f9663db), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				ohQ5V54TSgCjDaULgBN.qOys5bJ94(this.PSize, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x328CF016 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_1599679ee1724b48bf7a40ba1cf1dff5), ohQ5V54TSgCjDaULgBN.WZX4nJ3eZw);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PSize, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PSize, new System.Drawing.Size(140, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PSize, 243, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.PWord, ZYing.UI.FlowAlignment.Justify, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PWord, new System.Drawing.Point(157, 111), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				jxjTTThDV9tTypDLsCq.qOys5bJ94(this.PWord, new System.Drawing.Size(286, 0), jxjTTThDV9tTypDLsCq.fPshRMocUh);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PWord, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3034E68D ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_7a1fed405f4748f2b7b141bf9600fe2e), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				num = 1;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9b19501daa664a5590dccb8116e5c8fd == 0)
				{
					num = 1;
				}
				break;
			case 19:
				IrPSHnalZBMINCakkJt.qOys5bJ94(this.main, false, IrPSHnalZBMINCakkJt.QCWaFfWjhc);
				iY7ddqa4Gm4PwIUD571.qOys5bJ94(this.main, iY7ddqa4Gm4PwIUD571.Mw7amI9MqQ);
				IrPSHnalZBMINCakkJt.qOys5bJ94(this.lay多选, false, IrPSHnalZBMINCakkJt.QCWaFfWjhc);
				iY7ddqa4Gm4PwIUD571.qOys5bJ94(this.lay多选, iY7ddqa4Gm4PwIUD571.Mw7amI9MqQ);
				R1SI88aQ1uSr6F4NPe4.qOys5bJ94(this, false, R1SI88aQ1uSr6F4NPe4.xmxak2Irsb);
				return;
			case 20:
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.btn多选导入, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x30457451 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b6a459225bf544d6b93d37e0957a79c2), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.btn状态, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				num = 22;
				break;
			case 7:
				this.PTurn = new ZYing.UI.ETurn();
				this.lblTotal = new ZYing.UI.ELoader();
				num = 33;
				break;
			case 9:
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PLevel, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2D21F7B9 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ae94be659ed84b8b8b4b11eef55c3fd4), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PLevel, new System.Windows.Forms.Padding(0, 2, 0, 2), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PLevel, new System.Drawing.Size(147, 23), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PLevel, 8, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PGrade, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PGrade, new System.Drawing.Point(367, 178), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.PGrade, new System.Windows.Forms.Padding(5, 0, 25, 1), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PGrade, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-(-326985204) ^ 0x7A6E8113 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_88a9b435230c4eaf827b56a79d6ee53c), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				num = 12;
				break;
			case 35:
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.btn状态, new System.Drawing.Point(238, 566), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.btn状态, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x36B07A7D ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_12046703b6504e42927f2daf3fea0cdd), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.btn状态, new System.Windows.Forms.Padding(10, 3, 10, 3), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.btn状态, new System.Drawing.Size(48, 25), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.btn状态, 9, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.btn状态, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4CD13FE9 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_282f1df2aaef433d8ead5f2d044ac539), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				num = 18;
				break;
			case 33:
				this.btn导入主库 = new ZYing.UI.EButton();
				this.btn多选导入 = new ZYing.UI.EButton();
				this.btn状态 = new ZYing.UI.EButton();
				this.btn分类 = new ZYing.UI.EButton();
				this.btn删除 = new ZYing.UI.EButton();
				this.lay多选 = new ZYing.UI.ELayerAuto();
				this.chkSelect = new ZYing.UI.ECheck();
				this.lblAll = new ZYing.UI.ELabel();
				this.selOP = new ZYing.UI.ESelect();
				this.btnGO = new ZYing.UI.EButton();
				eaxvqahIffJWBdofyB3.qOys5bJ94(this.main, eaxvqahIffJWBdofyB3.ODrh2FqsFZ);
				eaxvqahIffJWBdofyB3.qOys5bJ94(this.lay多选, eaxvqahIffJWBdofyB3.ODrh2FqsFZ);
				OeB11Ehy2Eeerq75ANi.qOys5bJ94(this, OeB11Ehy2Eeerq75ANi.Jwohhmc6LB);
				FNN8aRITflH5QKGmHYLf.qOys5bJ94(this.right, new System.Windows.Forms.Padding(6, 4, 8, 4), FNN8aRITflH5QKGmHYLf.AMcITCTtpAA);
				KaqdrdITVBY92WxV7DuW.qOys5bJ94(this.right, 530, KaqdrdITVBY92WxV7DuW.FhXITdwkfP7);
				klRppNF8NYn2EiGjv1u.qOys5bJ94(this.right, System.Windows.Forms.DockStyle.Right, klRppNF8NYn2EiGjv1u.GsaFYi7gAS);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.right, new System.Drawing.Point(861, 0), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				num = 21;
				break;
			case 26:
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.btn多选导入, new System.Windows.Forms.Padding(10, 3, 10, 3), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.btn多选导入, new System.Drawing.Size(104, 25), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.btn多选导入, 265, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				num = 20;
				break;
			case 27:
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(ZYing.Channel.采集列表.TuYi1byrUlTo12YrfYwH(this.main), this.btn导入主库, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.btn多选导入, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.btn状态, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.btn分类, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.btn删除, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				num = 36;
				break;
			case 16:
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PLogin, 242, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				AIEQImtl6o6GfJuhUOe.qOys5bJ94(this.PLogin, false, AIEQImtl6o6GfJuhUOe.OVEtFTn5qP);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.Price, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				vQEZFs1ZMxsGmYwB0IJ.qOys5bJ94(this.Price, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x17B2795D ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_443e3241874d456eafc415ea3ca1dc23), vQEZFs1ZMxsGmYwB0IJ.CM11BOfRa6);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.Price, new System.Drawing.Point(709, 9), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.Price, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7F5E8282 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_71b15132f4c3474aaebfe1530c68df31), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.Price, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				aKEv471qMjkBlvVaum6.qOys5bJ94(this.Price, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1FAD3D22 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_f4abe29a753845b9a9a0f21644ea7e67), aKEv471qMjkBlvVaum6.YQO1vLDVvh);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.Price, new System.Drawing.Size(140, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.Price, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				tGysDe1tUAs33fKoiP9.qOys5bJ94(this.Price, ZYing.UI.EValidType.Price, tGysDe1tUAs33fKoiP9.qlv1sPDiqt);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.txtDelays, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				vQEZFs1ZMxsGmYwB0IJ.qOys5bJ94(this.txtDelays, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x526FA4DC ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_064ce8ea335043af8ca584aaced88385), vQEZFs1ZMxsGmYwB0IJ.CM11BOfRa6);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.txtDelays, new System.Drawing.Point(11, 43), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.txtDelays, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x279E8910 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9dda8de78a0a4b1983aabe885326adb3), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.txtDelays, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				aKEv471qMjkBlvVaum6.qOys5bJ94(this.txtDelays, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-(-326985204) ^ 0xCD058EA ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_f4abe29a753845b9a9a0f21644ea7e67), aKEv471qMjkBlvVaum6.YQO1vLDVvh);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.txtDelays, new System.Drawing.Size(140, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.txtDelays, 8, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				tGysDe1tUAs33fKoiP9.qOys5bJ94(this.txtDelays, ZYing.UI.EValidType.Number, tGysDe1tUAs33fKoiP9.qlv1sPDiqt);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PDate, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PDate, new System.Drawing.Point(157, 43), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PDate, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2CB2F491 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0efbc81bc23c42efac0295bf66972dea), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				num = 3;
				break;
			case 8:
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.PFlag, new System.Windows.Forms.Padding(5, 0, 25, 1), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PFlag, (string)ZYing.Channel.采集列表.AR4Zasyr5NWWQ9bPNWpT(0x7F5D204A ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_71b15132f4c3474aaebfe1530c68df31), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PFlag, new System.Windows.Forms.Padding(0, 2, 0, 2), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PFlag, new System.Drawing.Size(147, 25), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PFlag, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				num = 23;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_450309df683f4e4d8ed93c0e244cef4f == 0)
				{
					num = 17;
				}
				break;
			case 11:
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PType, new System.Drawing.Point(543, 178), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.PType, new System.Windows.Forms.Padding(5, 0, 25, 1), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PType, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2CB232B9 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0efbc81bc23c42efac0295bf66972dea), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PType, new System.Windows.Forms.Padding(0, 2, 0, 2), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PType, new System.Drawing.Size(150, 23), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PType, 4, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.grid, ZYing.UI.FlowAlignment.Fill, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.grid, new System.Drawing.Point(13, 209), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.grid, new System.Windows.Forms.Padding(5, 5, 3, 3), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				CJiQcbvTVov0V6nEgKj.qOys5bJ94(this.grid, 309, CJiQcbvTVov0V6nEgKj.RbmvnMsrU7);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.grid, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2D264DBD ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ae94be659ed84b8b8b4b11eef55c3fd4), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.grid, new System.Windows.Forms.Padding(10), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				Qw9jC1ITHWsCwqvA3oNj.qOys5bJ94(this.grid, false, Qw9jC1ITHWsCwqvA3oNj.er2ITiRaVXc);
				num = 32;
				break;
			case 23:
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PLevel, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PLevel, new System.Drawing.Point(190, 178), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.PLevel, new System.Windows.Forms.Padding(5, 0, 25, 1), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				num = 9;
				break;
			case 2:
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PSize, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PWord, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				num2 = 15;
				goto IL_000c;
			case 4:
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PArea, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x17B7B039 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_443e3241874d456eafc415ea3ca1dc23), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PArea, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				fp6d6cwfVDaI10P0G06.qOys5bJ94(this.PArea, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x19F51DE2 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_de4c858d184a45e7a9ea91377a1f9cba), fp6d6cwfVDaI10P0G06.crPwCHGS93);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PArea, new System.Drawing.Size(140, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				num = 12;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_01ceb1fa80cc4c7092dd54bea08b0961 == 0)
				{
					num = 25;
				}
				break;
			case 28:
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PBan, new System.Drawing.Size(77, 26), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PBan, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.PBan, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x142612D4 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_888e5b7a9cff40dea01c8d4cddb988ed), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				num = 13;
				break;
			case 36:
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.lay多选, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.selOP, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.btnGO, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				klRppNF8NYn2EiGjv1u.qOys5bJ94(this.main, System.Windows.Forms.DockStyle.Fill, klRppNF8NYn2EiGjv1u.GsaFYi7gAS);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.main, new System.Drawing.Point(0, 0), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.main, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1D16E6C5 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_8363360abdc34f81bce948eb97479daa), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.main, new System.Windows.Forms.Padding(8, 6, 8, 15), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.main, new System.Drawing.Size(861, 613), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.main, 6, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PKey, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PKey, new System.Drawing.Point(11, 9), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				cYfn1D4HCYpIA37B0Tb.qOys5bJ94(this.PKey, 3000, cYfn1D4HCYpIA37B0Tb.bGr4iojMLj);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PKey, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x65D909AA ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_7d77dc0ef27547e4ab0c2f1b6c3778ad), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PKey, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				BhIJPQ451xVVQN7PQWQ.qOys5bJ94(this.PKey, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7FB8FEF5 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_603178892e9e413b98f3e6aa3da54b45), BhIJPQ451xVVQN7PQWQ.koi4UpNTKR);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PKey, new System.Drawing.Size(140, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PKey, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				ksIErdheSQxyf1JPuEj.qOys5bJ94(this.PKey, ZYing.UI.EValidType.EN, ksIErdheSQxyf1JPuEj.q6Vh9lkrDF);
				wH28it4fIBHr55P7oYo.qOys5bJ94(this.PKind, true, wH28it4fIBHr55P7oYo.fBL4C8jHGd);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.PKind, ZYing.UI.FlowAlignment.Justify, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PKind, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PKind, new System.Drawing.Point(157, 9), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				weLAjHaq7jlOr3ysrIs.qOys5bJ94(this.PKind, new System.Drawing.Size(500, 28), weLAjHaq7jlOr3ysrIs.QqZavhlOQI);
				jxjTTThDV9tTypDLsCq.qOys5bJ94(this.PKind, new System.Drawing.Size(286, 0), jxjTTThDV9tTypDLsCq.fPshRMocUh);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PKind, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x11FF7A64 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9cd5a0aaeba04bb4b081b4b5fe90a190), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PKind, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				fp6d6cwfVDaI10P0G06.qOys5bJ94(this.PKind, (string)ZYing.Channel.采集列表.AR4Zasyr5NWWQ9bPNWpT(-(-326985204) ^ 0x6CC488F1 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_603178892e9e413b98f3e6aa3da54b45), fp6d6cwfVDaI10P0G06.crPwCHGS93);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PKind, new System.Drawing.Size(400, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				num = 29;
				break;
			default:
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.right, new System.Drawing.Size(550, 613), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.right, 6, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				fkA4S6IjI7LSIbj4rSeV.qOys5bJ94(this.right, new System.Windows.Forms.Padding(6, 9, 6, 6), fkA4S6IjI7LSIbj4rSeV.MNGIj23eWEa);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.right, (string)ZYing.Channel.采集列表.AR4Zasyr5NWWQ9bPNWpT(0x50846F47 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_fca4b5827c154beda800a45f1a8c082c), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				UbwBHZIja4iN7oTFHnuP.qOys5bJ94(this.right, false, UbwBHZIja4iN7oTFHnuP.xtcIjNDtxqA);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PKey, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PKind, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PLogin, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.Price, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.txtDelays, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PDate, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PArea, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PSource, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				num = 2;
				break;
			case 31:
				D4NRwwqppyKWqQ8lElg.qOys5bJ94(this.lblTotal, System.Windows.Forms.HorizontalAlignment.Right, D4NRwwqppyKWqQ8lElg.symq1wO9od);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.btn导入主库, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.btn导入主库, new System.Drawing.Point(11, 566), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.btn导入主库, new System.Windows.Forms.Padding(3, 3, 10, 3), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.btn导入主库, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6913FF29 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_88a9b435230c4eaf827b56a79d6ee53c), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.btn导入主库, new System.Windows.Forms.Padding(10, 3, 10, 3), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.btn导入主库, new System.Drawing.Size(104, 25), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.btn导入主库, 264, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.btn导入主库, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x17B279FD ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_443e3241874d456eafc415ea3ca1dc23), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.btn多选导入, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				pxHVUshl48ahyalcwFt.qOys5bJ94(this.btn多选导入, isFFjwkqk3pcZxCGKXY.qOys5bJ94(isFFjwkqk3pcZxCGKXY.moRkvPHrJR), pxHVUshl48ahyalcwFt.pgchF9iJkR);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.btn多选导入, new System.Drawing.Point(128, 566), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.btn多选导入, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x75CCEB8 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d77468cda0b945e3a4061d84efc82d59), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				num = 26;
				break;
			case 25:
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PArea, 491, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				AIEQImtl6o6GfJuhUOe.qOys5bJ94(this.PArea, false, AIEQImtl6o6GfJuhUOe.OVEtFTn5qP);
				num = 37;
				break;
			case 13:
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PFlag, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PFlag, new System.Drawing.Point(13, 178), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				num = 8;
				break;
			case 14:
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.btnGO, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x180370A5 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_fcd2f31ec7e946f5991ef1cf55616ab6), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.btnGO, new System.Windows.Forms.Padding(6, 4, 6, 2), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.btnGO, new System.Drawing.Size(33, 25), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.btnGO, 8, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.btnGO, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x26BA4E25 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_e6be942033a74c0baffa094b080e9e6b), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(base.Controls, this.main, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(base.Controls, this.right, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this, new System.Windows.Forms.Padding(3, 4, 3, 4), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				aYKORgy57UksQMRX1Ay.qOys5bJ94(this, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x381F7CA0 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9bf27d58a5584b2cb3d10730eccbd9c5), aYKORgy57UksQMRX1Ay.QnjyUtlCph);
				TtWehCqDGvRP7kwOYXW.qOys5bJ94(this, new System.Drawing.Size(1411, 613), TtWehCqDGvRP7kwOYXW.hGxqRFFBdX);
				num = 14;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_7d77dc0ef27547e4ab0c2f1b6c3778ad != 0)
				{
					num = 19;
				}
				break;
			case 32:
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.grid, new System.Drawing.Size(836, 309), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.grid, 8, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.grid, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-(-742678047) ^ 0x2A27199 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0f5d594c7dcb4a7f9e6cc9d946db1eec), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.PTurn, ZYing.UI.FlowAlignment.Justify, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PTurn, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PTurn, new System.Drawing.Point(11, 526), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.PTurn, new System.Windows.Forms.Padding(3, 5, 3, 10), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PTurn, (string)ZYing.Channel.采集列表.AR4Zasyr5NWWQ9bPNWpT(0x485469CB ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9e7c8ea59e0745258e93689531c631f8), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PTurn, new System.Windows.Forms.Padding(4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PTurn, new System.Drawing.Size(687, 27), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				num = 14;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_10ffe2be534f4959b052b27cc6e82b39 != 0)
				{
					num = 17;
				}
				break;
			case 12:
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PGrade, new System.Windows.Forms.Padding(0, 2, 0, 2), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PGrade, new System.Drawing.Size(146, 25), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PGrade, 3, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				num = 11;
				break;
			case 15:
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.eLine1, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PStat, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PBan, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PFlag, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				num2 = 24;
				goto IL_000c;
			case 21:
				weLAjHaq7jlOr3ysrIs.qOys5bJ94(this.right, new System.Drawing.Size(1500, 500000), weLAjHaq7jlOr3ysrIs.QqZavhlOQI);
				jxjTTThDV9tTypDLsCq.qOys5bJ94(this.right, new System.Drawing.Size(550, 0), jxjTTThDV9tTypDLsCq.fPshRMocUh);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.right, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7C23425E ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_472ad9b945424beebea66cf8ae5e521c), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.right, new System.Windows.Forms.Padding(1, 0, 1, 1), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				t1S8XsIT3pUDkQ5NE0Fm.qOys5bJ94(this.right, -1, t1S8XsIT3pUDkQ5NE0Fm.Wj9ITGKAsn6);
				num = 0;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_24274e778cb34f0b9c6e5b5bbe7cc59b == 0)
				{
					num = 0;
				}
				break;
			case 34:
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.lblTotal, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x70D95A2E ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_82991c292d16419fb0eb42340f9663db), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				Y42loFn3f7dRGCwOgL3.qOys5bJ94(this.lblTotal, new System.Windows.Forms.Padding(6, 4, 0, 4), Y42loFn3f7dRGCwOgL3.koynG9gHg7);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.lblTotal, new System.Drawing.Size(145, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.lblTotal, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.lblTotal, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x279E6A22 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9dda8de78a0a4b1983aabe885326adb3), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				num = 4;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_39b773881e74450ab394fb512634f040 == 0)
				{
					num = 31;
				}
				break;
			case 24:
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PLevel, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PGrade, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PType, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				num2 = 5;
				goto IL_000c;
			case 1:
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PWord, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				num = 30;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_5428c2381b124759bf53b9f3f7b3b5ec != 0)
				{
					num = 2;
				}
				break;
			case 3:
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PDate, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PDate, new System.Drawing.Size(286, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PDate, 9, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				wH28it4fIBHr55P7oYo.qOys5bJ94(this.PArea, true, wH28it4fIBHr55P7oYo.fBL4C8jHGd);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PArea, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PArea, new System.Drawing.Point(449, 43), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				num2 = 4;
				goto IL_000c;
			case 18:
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.btn分类, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				pxHVUshl48ahyalcwFt.qOys5bJ94(this.btn分类, jMlgugInlXJFOGfab0f4.qOys5bJ94(jMlgugInlXJFOGfab0f4.g2kInFLCST4), pxHVUshl48ahyalcwFt.pgchF9iJkR);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.btn分类, new System.Drawing.Point(292, 566), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.btn分类, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x3A4393EA ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_c0fff80a6c6a4ab1825b648985b91c3a), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.btn分类, new System.Windows.Forms.Padding(10, 3, 10, 3), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.btn分类, new System.Drawing.Size(48, 25), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.btn分类, 10, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.btn分类, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7BFCA4EB ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ea7ba9c275744d9da2d713459cb8f164), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.btn删除, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				pxHVUshl48ahyalcwFt.qOys5bJ94(this.btn删除, MvDIhQhf8kkmvZ4Uerf.qOys5bJ94(MvDIhQhf8kkmvZ4Uerf.SwohCFdEsQ), pxHVUshl48ahyalcwFt.pgchF9iJkR);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.btn删除, new System.Drawing.Point(346, 566), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.btn删除, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x48553AF5 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9e7c8ea59e0745258e93689531c631f8), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.btn删除, new System.Windows.Forms.Padding(10, 3, 10, 3), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.btn删除, new System.Drawing.Size(48, 25), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.btn删除, 13, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.btn删除, (string)ZYing.Channel.采集列表.AR4Zasyr5NWWQ9bPNWpT(0x381F2B4E ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9bf27d58a5584b2cb3d10730eccbd9c5), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				AIEQImtl6o6GfJuhUOe.qOys5bJ94(this.btn删除, false, AIEQImtl6o6GfJuhUOe.OVEtFTn5qP);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.lay多选.Controls, this.chkSelect, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.lay多选.Controls, this.lblAll, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.lay多选, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				num2 = 6;
				goto IL_000c;
			case 5:
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.grid, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.PTurn, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				AqUxBfyjckHiNpr2XWr.qOys5bJ94(this.main.Controls, this.lblTotal, AqUxBfyjckHiNpr2XWr.jdbybagiS6);
				num = 27;
				break;
			case 30:
				BhIJPQ451xVVQN7PQWQ.qOys5bJ94(this.PWord, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-1506365678 ^ -1641184360 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_9bf27d58a5584b2cb3d10730eccbd9c5), BhIJPQ451xVVQN7PQWQ.koi4UpNTKR);
				L5G1DEC4hrypf17LHoc.qOys5bJ94(this.PWord, true, L5G1DEC4hrypf17LHoc.VvkCmmBFpu);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PWord, new System.Drawing.Size(692, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PWord, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.eLine1, new System.Drawing.Point(11, 145), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.eLine1, new System.Windows.Forms.Padding(3, 3, 3, 5), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.eLine1, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x15584089 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_24274e778cb34f0b9c6e5b5bbe7cc59b), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.eLine1, new System.Drawing.Size(838, 2), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.PStat, ZYing.UI.FlowAlignment.Justify, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PStat, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PStat, new System.Drawing.Point(13, 152), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.PStat, new System.Windows.Forms.Padding(5, 0, 10, 1), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				jxjTTThDV9tTypDLsCq.qOys5bJ94(this.PStat, new System.Drawing.Size(200, 0), jxjTTThDV9tTypDLsCq.fPshRMocUh);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PStat, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x18045083 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_fcd2f31ec7e946f5991ef1cf55616ab6), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PStat, new System.Windows.Forms.Padding(7, 3, 7, 3), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PStat, new System.Drawing.Size(746, 25), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PStat, 1, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.PBan, ZYing.UI.FlowAlignment.Right, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PBan, new System.Drawing.Point(772, 152), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				num2 = 10;
				goto IL_000c;
			case 29:
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.PKind, 22, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				wH28it4fIBHr55P7oYo.qOys5bJ94(this.PLogin, true, wH28it4fIBHr55P7oYo.fBL4C8jHGd);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.PLogin, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				rNWWPW5lgPqTMYwqxNr.qOys5bJ94(this.PLogin, true, rNWWPW5lgPqTMYwqxNr.wqK5FBg3ar);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.PLogin, new System.Drawing.Point(563, 9), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PLogin, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4F8C7C14 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b480309f51bd48c194e1ed3315e9d650), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PLogin, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				fp6d6cwfVDaI10P0G06.qOys5bJ94(this.PLogin, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x526B79BA ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_064ce8ea335043af8ca584aaced88385), fp6d6cwfVDaI10P0G06.crPwCHGS93);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.PLogin, new System.Drawing.Size(140, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				num = 16;
				break;
			case 6:
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.lay多选, new System.Drawing.Point(400, 566), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.lay多选, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x6913E423 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_88a9b435230c4eaf827b56a79d6ee53c), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.lay多选, new System.Drawing.Size(89, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.lay多选, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.lay多选, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x4CD2CEA1 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_282f1df2aaef433d8ead5f2d044ac539), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.chkSelect, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				tithwxnV248BrJgXeZC.qOys5bJ94(this.chkSelect, false, tithwxnV248BrJgXeZC.ONZndEnyJ5);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.chkSelect, new System.Drawing.Point(0, 0), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.chkSelect, new System.Windows.Forms.Padding(0), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.chkSelect, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x581B14DB ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_ed88ade44dea498194ae1c8221333918), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.chkSelect, new System.Windows.Forms.Padding(6, 4, 6, 4), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.chkSelect, new System.Drawing.Size(57, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.chkSelect, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.chkSelect, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x41B9DCC0 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_141040ccb76546af8ef08c87257c6ac9), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.lblAll, new System.Drawing.Point(60, 0), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.lblAll, new System.Windows.Forms.Padding(3, 0, 0, 0), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.lblAll, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x7807CD91 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_488581fc99e442088402176de3de7cbd), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				fsnwqEvVIF4wNEmb0o0.qOys5bJ94(this.lblAll, new System.Windows.Forms.Padding(0, 4, 0, 4), fsnwqEvVIF4wNEmb0o0.icRvd21IPc);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.lblAll, new System.Drawing.Size(29, 28), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.lblAll, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				jLTKnh2z5j5GYdWFTT5.qOys5bJ94(this.lblAll, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x2EE7CEF8 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_0f5d594c7dcb4a7f9e6cc9d946db1eec), jLTKnh2z5j5GYdWFTT5.bZKyOhr3yw);
				wH28it4fIBHr55P7oYo.qOys5bJ94(this.selOP, true, wH28it4fIBHr55P7oYo.fBL4C8jHGd);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.selOP, ZYing.UI.FlowAlignment.Right, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				XjDC0BhPw7vcOJrDFtf.qOys5bJ94(this.selOP, false, XjDC0BhPw7vcOJrDFtf.fWmhuneMOm);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.selOP, new System.Drawing.Point(725, 566), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				MmFOZehQncDxKqg8KE8.qOys5bJ94(this.selOP, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x173AE0B ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_dfe08b55d34746fe8ba0141c88660b08), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
				ohQ5V54TSgCjDaULgBN.qOys5bJ94(this.selOP, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x74EDB3F2 ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_d59bc24a307e4bdca17ad3010cacc7b5), ohQ5V54TSgCjDaULgBN.WZX4nJ3eZw);
				CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.selOP, new System.Windows.Forms.Padding(6, 3, 6, 3), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
				peRpaMhZvA9jHK1U9Yr.qOys5bJ94(this.selOP, new System.Drawing.Size(85, 26), peRpaMhZvA9jHK1U9Yr.VT1hBegrVS);
				u1vXWHhq3rpfB63eo2K.qOys5bJ94(this.selOP, 7, u1vXWHhq3rpfB63eo2K.HFJhvYAvY4);
				nDlDE7habqlortZg9Pi.qOys5bJ94(this.btnGO, ZYing.UI.FlowAlignment.Right, nDlDE7habqlortZg9Pi.FcMhN7AeLQ);
				pxHVUshl48ahyalcwFt.qOys5bJ94(this.btnGO, aQHLxQhKt15RVtJiKfo.qOys5bJ94(aQHLxQhKt15RVtJiKfo.PAchLoDpQ8), pxHVUshl48ahyalcwFt.pgchF9iJkR);
				oX2517h4kFNBQNRV21T.qOys5bJ94(this.btnGO, new System.Drawing.Point(816, 566), oX2517h4kFNBQNRV21T.O6mhm4cPl4);
				num = 14;
				if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_b227bbab40e0440284ca7e684c696e92 != 0)
				{
					num = 4;
				}
				break;
			case 22:
				pxHVUshl48ahyalcwFt.qOys5bJ94(this.btn状态, aQHLxQhKt15RVtJiKfo.qOys5bJ94(aQHLxQhKt15RVtJiKfo.PAchLoDpQ8), pxHVUshl48ahyalcwFt.pgchF9iJkR);
				num = 35;
				break;
			case 10:
				{
					BNO3oVhHEhjAvtd1WxH.qOys5bJ94(this.PBan, new System.Windows.Forms.Padding(3, 0, 3, 1), BNO3oVhHEhjAvtd1WxH.Nfhhi6hrYn);
					MmFOZehQncDxKqg8KE8.qOys5bJ94(this.PBan, Hmg1hUwjXmJQ0hx3wd.A8RedEce1ZSTaHmUDS.DQmUrYZ9D(-(-742678047) ^ 0x2F4573FF ^ <Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_1dfe6a356d324a4aba1c21d85b6856ea), MmFOZehQncDxKqg8KE8.tBqhk8Y9Yi);
					CdPHW6hTcJmT2r101Z6.qOys5bJ94(this.PBan, new System.Windows.Forms.Padding(0, 3, 6, 3), CdPHW6hTcJmT2r101Z6.vsKhn7u2Kw);
					num = 28;
					if (<Module>{7d0facdd-f73e-4b8b-a190-3fe5c2bba644}.m_d94cb8ffb8b4419e93c1a1ae001cd345.m_944c85cf5b3d495ea454c6ac0f9a3fa4 == 0)
					{
						num = 21;
					}
					break;
				}
				IL_000c:
				num = num2;
				break;
			}
		}
	}

	static 采集列表()
	{
		A8RedEce1ZSTaHmUDS.iR3aCLEY7();
	}

	internal static bool L8YeHVyrxhyL8OPoAl1T()
	{
		return PWyTwPyrM2qPxaMNINe3 == null;
	}

	internal static 采集列表 opxaXhyrTLZJQN8lpAlA()
	{
		return PWyTwPyrM2qPxaMNINe3;
	}

	internal static void zFdO2SyrnGnaLgwtHY2U(object P_0, object P_1)
	{
		((EGallery)P_0).CellCreating += (EGallery.CreateHandler)P_1;
	}

	internal static object AR4Zasyr5NWWQ9bPNWpT(int P_0)
	{
		return A8RedEce1ZSTaHmUDS.DQmUrYZ9D(P_0);
	}

	internal static object TuYi1byrUlTo12YrfYwH(object P_0)
	{
		return ((EControl)P_0).Controls;
	}
}
