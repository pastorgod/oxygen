package xnet

import ()

type Context struct {
	Conn          ISession // 本次会话
	packet        *Packet  // 本次会话包
	asynchronized bool     // 是否转换为异步模式
}

func NewContext(conn ISession, packet *Packet) *Context {
	return &Context{Conn: conn, packet: packet}
}

// 返回回应数据
func (this *Context) Response(err *string, msg Message) {
	this.Conn.Response(err, msg, this.packet.RequestId())
}

// 转换为异步模式
func (this *Context) Asynchronized() *string {
	this.asynchronized = true
	return nil
}

// 是否异步模式
func (this *Context) isAsynced() bool {
	return this.asynchronized
}

/*

++IapService.Buy( 100 )
+  Trace {
+	trace_id : xxx-111-222-33-fff-aaa,
+	service: IapService,
+	method: Buy,
+	Span: {
+		parent: 0,
+		span_id: 1,
+		begin_time: 1359809800,
+		end_time: 135980920,
+		req_bytes: 4,
+		rsp_bytes: 35,
+		status: ok,
+	},
+  }
+
+  +++TradeCenter.CreateOrder( user-1222341, 100 )
+		Trace {
+			trace_id: xxx-111-222-33-fff-aaa,
+			service: TradeCenter,
+			method: CreateOrder,
+			Span: {
+				parent: 1,
+				span_id: 2,
+				begin_time: 1359809810,
+				end_time: 135980913,
+				req_bytes: 14,
+				rsp_bytes: 31,
+				status: ok,
+			},
+		}
+
+
+	   NotifyCenter.Notify( user-1222341, 201510121123412 )
+			Trace {
+				trace_id: xxx-111-222-33-fff-aaa,
+				service: NotifyCenter,
+				method: Notify,
+				Span: {
+					parent: 2,
+					span_id: 3,
+					begin_time: 1359809811,
+					end_time: 135980912,
+					req_bytes: 4,
+					rsp_bytes: 1,
+					status: ok,
+				},
+			}

	UnipayCenter.Pay( 10, 201510121123412 )
		Trace {
			trace_id: xxx-111-222-33-fff-aaa,
			service: UnipayCenter,
			method: Pay,
			Span: {
				parent: 1,
				span_id: 3,
				begin_time: 1359809811,
				end_time: 135980912,
				req_bytes: 22,
				rsp_bytes: 87,
				status: ok,
			},
		}

*/
