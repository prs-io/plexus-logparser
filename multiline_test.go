package logparser

import (
	"context"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeByLine(m *MultilineCollector, data string, ts time.Time) []Message {
	var msgs []Message
	done := make(chan bool)
	go func() {
		timer := time.NewTimer(3 * m.timeout)
		for {
			select {
			case <-timer.C:
				done <- true
				return
			case msg := <-m.Messages:
				msgs = append(msgs, msg)
			}
		}
	}()
	for _, line := range strings.Split(data, "\n") {
		m.Add(LogEntry{Timestamp: ts, Content: line, Level: LevelUnknown})
		ts = ts.Add(time.Millisecond)
	}
	<-done
	return msgs
}

func TestMultilineCollector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewMultilineCollector(ctx, 10*time.Millisecond, multilineCollectorLimit)
	defer cancel()

	data := `Order response: {"statusCode":406,"body":{"timestamp":1648205755430,"status":406,"error":"Not Acceptable","exception":"works.weave.socks.orders.controllers.OrdersController$PaymentDeclinedException","message":"Payment declined: amount exceeds 100.00","path":"/orders"},"headers":{"x-application-context":"orders:80","content-type":"application/json;charset=UTF-8","transfer-encoding":"chunked","date":"Fri, 25 Mar 2022 10:55:55 GMT","connection":"close"},"request":{"uri":{"protocol":"http:","slashes":true,"auth":null,"host":"orders","port":80,"hostname":"orders","hash":null,"search":null,"query":null,"pathname":"/orders","path":"/orders","href":"http://orders/orders"},"method":"POST","headers":{"accept":"application/json","content-type":"application/json","content-length":232}}}
Order response: {"timestamp":1648205755430,"status":406,"error":"Not Acceptable","exception":"works.weave.socks.orders.controllers.OrdersController$PaymentDeclinedException","message":"Payment declined: amount exceeds 100.00","path":"/orders"}`
	msgs := writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 2)
	assert.Equal(t, strings.Split(data, "\n")[0], msgs[0].Content)
	assert.Equal(t, strings.Split(data, "\n")[1], msgs[1].Content)
}

func TestMultilineCollectorPython(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewMultilineCollector(ctx, 10*time.Millisecond, multilineCollectorLimit)
	defer cancel()

	data := `Traceback (most recent call last):
  File "/Users/user/workspace/pythonProject/main.py", line 10, in <module>
    func()
  File "/Users/user/workspace/pythonProject/main.py", line 4, in func
    raise ConnectionError
ConnectionError`
	msgs := writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)

	data = `Traceback (most recent call last):
  File "/Users/user/workspace/pythonProject/main.py", line 10, in <module>
    func()
  File "/Users/user/workspace/pythonProject/main.py", line 4, in func
    raise ConnectionError
ConnectionError
Traceback (most recent call last):
  File "/Users/user/workspace/pythonProject/main.py", line 10, in <module>
    func()
  File "/Users/user/workspace/pythonProject/main.py", line 4, in func
    raise ConnectionError
ConnectionError`
	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 2)
	assert.Equal(t, data, msgs[0].Content+"\n"+msgs[0].Content)

	data = `Traceback (most recent call last):
  File "/Users/user/workspace/pythonProject/main.py", line 10, in <module>
    func()
  File "/Users/user/workspace/pythonProject/main.py", line 4, in func
    raise ConnectionError
ConnectionError

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/Users/user/workspace/pythonProject/main.py", line 12, in <module>
    raise RuntimeError('Failed to open database') from exc
RuntimeError: Failed to open database`
	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)

	data = `Traceback (most recent call last):
  File "/Users/user/workspace/pythonProject/main.py", line 10, in <module>
    func()
  File "/Users/user/workspace/pythonProject/main.py", line 4, in func
    raise ConnectionError
ConnectionError

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/Users/user/workspace/pythonProject/main.py", line 12, in <module>
    raise RuntimeError('Failed to open database') from exc
RuntimeError: Failed to open database

During handling of the above exception, another exception occurred:

Traceback (most recent call last):
  File "/Users/user/workspace/pythonProject/main.py", line 14, in <module>
    raise ConnectionError
ConnectionError`
	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)

	data = `2020-03-20 08:48:57,067 ERROR [django.request:222] log 46 140452532862280 Internal Server Error: /article
Traceback (most recent call last):
  File "/usr/local/lib/python3.8/site-packages/django/db/backends/base/base.py", line 220, in ensure_connection
    self.connect()
  File "/usr/local/lib/python3.8/site-packages/django/utils/asyncio.py", line 26, in inner
    return func(*args, **kwargs)
  File "/usr/local/lib/python3.8/site-packages/django/db/backends/base/base.py", line 197, in connect
    self.connection = self.get_new_connection(conn_params)
  File "/usr/local/lib/python3.8/site-packages/django_prometheus/db/common.py", line 44, in get_new_connection
    return super(DatabaseWrapperMixin, self).get_new_connection(*args, **kwargs)
  File "/usr/local/lib/python3.8/site-packages/django/utils/asyncio.py", line 26, in inner
    return func(*args, **kwargs)
  File "/usr/local/lib/python3.8/site-packages/django/db/backends/mysql/base.py", line 233, in get_new_connection
    return Database.connect(**conn_params)
  File "/usr/local/lib/python3.8/site-packages/MySQLdb/__init__.py", line 84, in Connect
    return Connection(*args, **kwargs)
  File "/usr/local/lib/python3.8/site-packages/MySQLdb/connections.py", line 179, in __init__
    super(Connection, self).__init__(*args, **kwargs2)
MySQLdb._exceptions.OperationalError: (1040, 'Too many connections')

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/usr/local/lib/python3.8/site-packages/django/core/handlers/exception.py", line 34, in inner
    response = get_response(request)
  File "/usr/local/lib/python3.8/site-packages/django/core/handlers/base.py", line 115, in _get_response
    response = self.process_exception_by_middleware(e, request)
  File "/usr/local/lib/python3.8/site-packages/django/core/handlers/base.py", line 113, in _get_response
    response = wrapped_callback(request, *callback_args, **callback_kwargs)
  File "/usr/local/lib/python3.8/contextlib.py", line 74, in inner
    with self._recreate_cm():
  File "/usr/local/lib/python3.8/site-packages/django/db/transaction.py", line 175, in __enter__
    if not connection.get_autocommit():
  File "/usr/local/lib/python3.8/site-packages/django/db/backends/base/base.py", line 390, in get_autocommit
    self.ensure_connection()
  File "/usr/local/lib/python3.8/site-packages/django/utils/asyncio.py", line 26, in inner
    return func(*args, **kwargs)
  File "/usr/local/lib/python3.8/site-packages/django/db/backends/base/base.py", line 220, in ensure_connection
    self.connect()
  File "/usr/local/lib/python3.8/site-packages/django/db/utils.py", line 90, in __exit__
    raise dj_exc_value.with_traceback(traceback) from exc_value
  File "/usr/local/lib/python3.8/site-packages/django/db/backends/base/base.py", line 220, in ensure_connection
    self.connect()
  File "/usr/local/lib/python3.8/site-packages/django/utils/asyncio.py", line 26, in inner
    return func(*args, **kwargs)"
  File "/usr/local/lib/python3.8/site-packages/django/db/backends/base/base.py", line 197, in connect
    self.connection = self.get_new_connection(conn_params)
  File "/usr/local/lib/python3.8/site-packages/django_prometheus/db/common.py", line 44, in get_new_connection
    return super(DatabaseWrapperMixin, self).get_new_connection(*args, **kwargs)
  File "/usr/local/lib/python3.8/site-packages/django/utils/asyncio.py", line 26, in inner
    return func(*args, **kwargs)
  File "/usr/local/lib/python3.8/site-packages/django/db/backends/mysql/base.py", line 233, in get_new_connection
    return Database.connect(**conn_params)
  File "/usr/local/lib/python3.8/site-packages/MySQLdb/__init__.py", line 84, in Connect
    return Connection(*args, **kwargs)
  File "/usr/local/lib/python3.8/site-packages/MySQLdb/connections.py", line 179, in __init__
    super(Connection, self).__init__(*args, **kwargs2)
django.db.utils.OperationalError: (1040, 'Too many connections')`

	msgs = writeByLine(m, data, time.Unix(100500, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)
	assert.Equal(t, time.Unix(100500, 0), msgs[0].Timestamp)

	data = `2020-03-20 08:48:57,067 ERROR:__main__:Traceback (most recent call last):
  File "<stdin>", line 2, in <module>
  File "<stdin>", line 2, in do_something_that_might_error
  File "<stdin>", line 2, in raise_error
RuntimeError: something bad happened!`
	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)
}

func TestMultilineCollectorJava(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewMultilineCollector(ctx, 10*time.Millisecond, multilineCollectorLimit)
	defer cancel()

	data := `Exception in thread "main" java.lang.NullPointerException
	at com.example.MyClass.methodA(MyClass.java:10)
	at com.example.MyClass.methodB(MyClass.java:20)
	at com.example.MyClass.main(MyClass.java:30)
Caused by: java.lang.ArrayIndexOutOfBoundsException: Index 5 out of bounds for length 5
	at com.example.AnotherClass.anotherMethod(AnotherClass.java:15)
	at com.example.MyClass.methodA(MyClass.java:8)
	... 2 more`
	msgs := writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)

	data = `Exception in thread "main" java.lang.NullPointerException
	at com.example.MyClass.methodA(MyClass.java:10)
	at com.example.MyClass.methodB(MyClass.java:20)
	at com.example.MyClass.main(MyClass.java:30)
Caused by: java.lang.ArrayIndexOutOfBoundsException: Index 5 out of bounds for length 5
	at com.example.AnotherClass.anotherMethod(AnotherClass.java:15)
	at com.example.MyClass.methodA(MyClass.java:8)
	... 2 more
Exception in thread "main" java.lang.NullPointerException
	at com.example.MyClass.methodA(MyClass.java:10)
	at com.example.MyClass.methodB(MyClass.java:20)
	at com.example.MyClass.main(MyClass.java:30)
Caused by: java.lang.ArrayIndexOutOfBoundsException: Index 5 out of bounds for length 5
	at com.example.AnotherClass.anotherMethod(AnotherClass.java:15)
	at com.example.MyClass.methodA(MyClass.java:8)
	... 2 more`
	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 2)
	assert.Equal(t, data, msgs[0].Content+"\n"+msgs[1].Content)

	data = `ERROR [Messaging-EventLoop-3-1] 2023-10-04 14:27:35,249 2020-03-31 javax.servlet.ServletException: Something bad happened
	at com.example.myproject.OpenSessionInViewFilter.doFilter(OpenSessionInViewFilter.java:60)
	at org.mortbay.jetty.servlet.ServletHandler$CachedChain.doFilter(ServletHandler.java:1157)
	at com.example.myproject.ExceptionHandlerFilter.doFilter(ExceptionHandlerFilter.java:28)
	at org.mortbay.jetty.servlet.ServletHandler$CachedChain.doFilter(ServletHandler.java:1157)
	at com.example.myproject.OutputBufferFilter.doFilter(OutputBufferFilter.java:33)
	at org.mortbay.jetty.servlet.ServletHandler$CachedChain.doFilter(ServletHandler.java:1157)
	at org.mortbay.jetty.servlet.ServletHandler.handle(ServletHandler.java:388)
	at org.mortbay.jetty.security.SecurityHandler.handle(SecurityHandler.java:216)
	at org.mortbay.jetty.servlet.SessionHandler.handle(SessionHandler.java:182)
	at org.mortbay.jetty.handler.ContextHandler.handle(ContextHandler.java:765)
	at org.mortbay.jetty.webapp.WebAppContext.handle(WebAppContext.java:418)
	at org.mortbay.jetty.handler.HandlerWrapper.handle(HandlerWrapper.java:152)
	at org.mortbay.jetty.Server.handle(Server.java:326)
	at org.mortbay.jetty.HttpConnection.handleRequest(HttpConnection.java:542)
	at org.mortbay.jetty.HttpConnection$RequestHandler.content(HttpConnection.java:943)
	at org.mortbay.jetty.HttpParser.parseNext(HttpParser.java:756)
	at org.mortbay.jetty.HttpParser.parseAvailable(HttpParser.java:218)
	at org.mortbay.jetty.HttpConnection.handle(HttpConnection.java:404)
	at org.mortbay.jetty.bio.SocketConnector$Connection.run(SocketConnector.java:228)
	at org.mortbay.thread.QueuedThreadPool$PoolThread.run(QueuedThreadPool.java:582)
Caused by: com.example.myproject.MyProjectServletException
	at com.example.myproject.MyServlet.doPost(MyServlet.java:169)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:727)
	at javax.servlet.http.HttpServlet.service(HttpServlet.java:820)
	at org.mortbay.jetty.servlet.ServletHolder.handle(ServletHolder.java:511)
	at org.mortbay.jetty.servlet.ServletHandler$CachedChain.doFilter(ServletHandler.java:1166)
	at com.example.myproject.OpenSessionInViewFilter.doFilter(OpenSessionInViewFilter.java:30)
	... 27 more
Caused by: org.hibernate.exception.ConstraintViolationException: could not insert: [com.example.myproject.MyEntity]
	at org.hibernate.exception.SQLStateConverter.convert(SQLStateConverter.java:96)
	at org.hibernate.exception.JDBCExceptionHelper.convert(JDBCExceptionHelper.java:66)
	at org.hibernate.id.insert.AbstractSelectingDelegate.performInsert(AbstractSelectingDelegate.java:64)
	at org.hibernate.persister.entity.AbstractEntityPersister.insert(AbstractEntityPersister.java:2329)
	at org.hibernate.persister.entity.AbstractEntityPersister.insert(AbstractEntityPersister.java:2822)
	at org.hibernate.action.EntityIdentityInsertAction.execute(EntityIdentityInsertAction.java:71)
	at org.hibernate.engine.ActionQueue.execute(ActionQueue.java:268)
	at org.hibernate.event.def.AbstractSaveEventListener.performSaveOrReplicate(AbstractSaveEventListener.java:321)
	at org.hibernate.event.def.AbstractSaveEventListener.performSave(AbstractSaveEventListener.java:204)
	at org.hibernate.event.def.AbstractSaveEventListener.saveWithGeneratedId(AbstractSaveEventListener.java:130)
	at org.hibernate.event.def.DefaultSaveOrUpdateEventListener.saveWithGeneratedOrRequestedId(DefaultSaveOrUpdateEventListener.java:210)
	at org.hibernate.event.def.DefaultSaveEventListener.saveWithGeneratedOrRequestedId(DefaultSaveEventListener.java:56)
	at org.hibernate.event.def.DefaultSaveOrUpdateEventListener.entityIsTransient(DefaultSaveOrUpdateEventListener.java:195)
	at org.hibernate.event.def.DefaultSaveEventListener.performSaveOrUpdate(DefaultSaveEventListener.java:50)
	at org.hibernate.event.def.DefaultSaveOrUpdateEventListener.onSaveOrUpdate(DefaultSaveOrUpdateEventListener.java:93)
	at org.hibernate.impl.SessionImpl.fireSave(SessionImpl.java:705)
	at org.hibernate.impl.SessionImpl.save(SessionImpl.java:693)
	at org.hibernate.impl.SessionImpl.save(SessionImpl.java:689)
	at sun.reflect.GeneratedMethodAccessor5.invoke(Unknown Source)
	at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:25)
	at java.lang.reflect.Method.invoke(Method.java:597)
	at org.hibernate.context.ThreadLocalSessionContext$TransactionProtectionWrapper.invoke(ThreadLocalSessionContext.java:344)
	at $Proxy19.save(Unknown Source)
	at com.example.myproject.MyEntityService.save(MyEntityService.java:59) <-- relevant call (see notes below)
	at com.example.myproject.MyServlet.doPost(MyServlet.java:164)
	... 32 more
Caused by: java.sql.SQLException: Violation of unique constraint MY_ENTITY_UK_1: duplicate value(s) for column(s) MY_COLUMN in statement [...]
	at org.hsqldb.jdbc.Util.throwError(Unknown Source)
	at org.hsqldb.jdbc.jdbcPreparedStatement.executeUpdate(Unknown Source)
	at com.mchange.v2.c3p0.impl.NewProxyPreparedStatement.executeUpdate(NewProxyPreparedStatement.java:105)
	at org.hibernate.id.insert.AbstractSelectingDelegate.performInsert(AbstractSelectingDelegate.java:57)
	... 54 more`

	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)
}

func TestMultilineCollectorJS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewMultilineCollector(ctx, 10*time.Millisecond, multilineCollectorLimit)
	defer cancel()

	data := `UnauthorizedException [Error]: jwt expired
    at AuthMixingGuard.canActivate (/app/dist/core/auth.guard.js:23:27)
    at GuardsConsumer.tryActivate (/app/node_modules/@nestjs/core/guards/guards-consumer.js:15:34)
    at canActivateFn (/app/node_modules/@nestjs/core/router/router-execution-context.js:134:59)
    at /app/node_modules/@nestjs/core/router/router-execution-context.js:42:37
    at AsyncLocalStorage.run (async_hooks.js:314:14)
    at /app/node_modules/@nestjs/core/router/router-proxy.js:9:23 {
  response: { statusCode: 401, message: 'jwt expired', error: 'Unauthorized' },`
	msgs := writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)

	data = `Error: Invalid IV length
    at Decipheriv.createCipherBase (internal/crypto/cipher.js:103:19)
    at Decipheriv.createCipherWithIV (internal/crypto/cipher.js:121:20)
    at new Decipheriv (internal/crypto/cipher.js:264:22)
    at Object.createDecipheriv (crypto.js:130:10)
    at AuthMixingGuard.canActivate (/app/dist/core/auth.guard.js:19:92)
    at GuardsConsumer.tryActivate (/app/node_modules/@nestjs/core/guards/guards-consumer.js:15:34)
    at canActivateFn (/app/node_modules/@nestjs/core/router/router-execution-context.js:134:59)
    at /app/node_modules/@nestjs/core/router/router-execution-context.js:42:37
::ffff:10.50.10.96 - - [15/Feb/2024:09:11:15 +0000] "POST /api/auth HTTP/1.1" 500 47 "https://example.com/" "Mozilla/5.0 ..."`
	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 2)
	assert.Equal(t, data, msgs[0].Content+"\n"+msgs[1].Content)

	data = `Error: 14 UNAVAILABLE: read ECONNRESET
    at callErrorFromStatus (/app/node_modules/@grpc/grpc-js/build/src/call.js:31:19)
    at Object.onReceiveStatus (/app/node_modules/@grpc/grpc-js/build/src/client.js:192:76)
    at Object.onReceiveStatus (/app/node_modules/@grpc/grpc-js/build/src/client-interceptors.js:360:141)
    at Object.onReceiveStatus (/app/node_modules/@grpc/grpc-js/build/src/client-interceptors.js:323:181)
    at /app/node_modules/@grpc/grpc-js/build/src/resolving-call.js:99:78
    at process.processTicksAndRejections (node:internal/process/task_queues:77:11)
for call at
    at ServiceClientImpl.makeUnaryRequest (/app/node_modules/@grpc/grpc-js/build/src/client.js:160:32)
    at ServiceClientImpl.<anonymous> (/app/node_modules/@grpc/grpc-js/build/src/make-client.js:105:19)
    at /app/node_modules/@opentelemetry/instrumentation-grpc/build/src/grpc-js/clientUtils.js:131:31
    at /app/node_modules/@opentelemetry/instrumentation-grpc/build/src/grpc-js/index.js:233:209
    at AsyncLocalStorage.run (node:async_hooks:338:14)
    at AsyncLocalStorageContextManager.with (/app/node_modules/@opentelemetry/context-async-hooks/build/src/AsyncLocalStorageContextManager.js:33:40)
    at ContextAPI.with (/app/node_modules/@opentelemetry/api/build/src/api/context.js:60:46)
    at ServiceClientImpl.clientMethodTrace [as getProduct] (/app/node_modules/@opentelemetry/instrumentation-grpc/build/src/grpc-js/index.js:233:42)
    at /app/.next/server/chunks/813.js:65:58
    at new Promise (<anonymous>) {
  code: 14,
  details: 'read ECONNRESET',
  metadata: Metadata { internalRepr: Map(0) {}, options: {} }
}`
	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)
}

func TestMultilineCollectorGO(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewMultilineCollector(ctx, 10*time.Millisecond, multilineCollectorLimit)
	defer cancel()

	// TODO: `panic` without timestamp in the first line

	data := ``
	msgs := writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, data, msgs[0].Content)
}
func TestMultilineCollectorLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := NewMultilineCollector(ctx, 10*time.Millisecond, 100)
	defer cancel()
	data := "I0215 12:33:07.230967 foo\n" + strings.Repeat("foo\n\n\n", 20)
	assert.Equal(t, 146, len(data))
	msgs := writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, 100, len(msgs[0].Content))

	data = "I0215 12:33:07.230967" + strings.Repeat(" foo", 25)
	assert.Equal(t, 121, len(data))
	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, 100, len(msgs[0].Content))

	data = "I0215 12:33:07.230967" + strings.Repeat(" €", 25)
	assert.Equal(t, 121, len(data))
	msgs = writeByLine(m, data, time.Unix(0, 0))
	require.Len(t, msgs, 1)
	assert.Equal(t, 97, len(msgs[0].Content))
	assert.True(t, utf8.ValidString(msgs[0].Content))
}
