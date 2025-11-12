import contextlib

@contextlib.contextmanager
def suppress(e):
	try:
		yield
	except e:
		pass
