query_api_proxy
===================

This is a simple http json-rpc proxy. The request will be forwarded to multiple configured worker nodes at the same time.
 Before the configured timeout period, the return of the worker node will be selected and most returned to the client.
 If there is a node that does not return in time or if the result returned by the node is inconsistent with the result returned by other nodes,
 the error log is recorded. The jsons allowed to be returned are different in sequence due to serialization.