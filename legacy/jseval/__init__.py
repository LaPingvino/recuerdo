#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2013-2014, 2017, Marten de Vries
#
#	This file is part of OpenTeacher.
#
#	OpenTeacher is free software: you can redistribute it and/or modify
#	it under the terms of the GNU General Public License as published by
#	the Free Software Foundation, either version 3 of the License, or
#	(at your option) any later version.
#
#	OpenTeacher is distributed in the hope that it will be useful,
#	but WITHOUT ANY WARRANTY; without even the implied warranty of
#	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#	GNU General Public License for more details.
#
#	You should have received a copy of the GNU General Public License
#	along with OpenTeacher.  If not, see <http://www.gnu.org/licenses/>.

from PyQt5 import QtCore, QtQml
import re
import utils
import collections
import datetime
import os
import exceptions
import traceback
import sys
import itertools
import warnings
import pyproxies

INIT_SCRIPT_NAME = "_js_eval_init.js"
INIT_SCRIPT = """(function () {
  // Sets up a system of function wrappers, because QtQml cannot directly insert
  // arbitrary functions into the runtime. Instead, a single callback function
  // is used, which is called with an id that is unique per function. Data, both
  // arguments and return type, are passed via a global variable
  //
  // buildJSWrapper creates a wrapper over all that. Note that for identical
  // ids, identical wrapper functions will be returned.

  var cache = {};
  var exports = this._callback = {};

  function newWrapper(id) {
    return function() {
      exports.args = Array.prototype.slice.call(arguments);

      exports.pyCallback(id);

      var result = exports.result;
      delete exports.result;
      if (result.type === 'result') {
        return result.value;
      } else {
        var err = new Error(result.value.message);
        err.name = result.value.name;
        err.oldTraceback = result.value.oldTraceback;
        throw err;
      }
    };
  }

  exports.buildJSWrapper = function (id) {
    if (!cache[id]) {
      cache[id] = newWrapper(id);
    }
    return cache[id];
  };
}());
"""

# Used by the except hook to filter out jseval stuff from the traceback.
# Change it to an empty list to temporarily disable filtering (e.g. for
# debugging)
JSEVAL_MASKED_FILES = [
    os.path.join(os.path.dirname(__file__), '__init__.py'),
    os.path.join(os.path.dirname(__file__), 'pyproxies.py'),
]

# Used to remember exception classes, so exceptions can be recreated
pythonExceptionStore = dict(
    (name, exc) for name, exc in vars(exceptions).iteritems()
    if type(exc) == type(Exception)
)

class Callback(QtCore.QObject):
    """A QObject that can be registered in QtQml to provide a callback for
    init.js

    """
    def __init__(self, callbackData):
        super(Callback, self).__init__()
        self.funcs = {}
        self.callbackData = callbackData

    @QtCore.pyqtSlot(unicode)
    def callback(self, identifier):
        func = self.funcs[identifier]

        args = list(self.callbackData.pop('args'))
        try:
            # try to use the last argument as kwargs
            result = func(*args[:-1], **dict(args[-1].iteritems()))
        except (AttributeError, TypeError, IndexError):
            try:
                result = func(*args)
            except BaseException, e:
                if not isinstance(e, pyproxies.JSError):
                    #store the exception class so it can be reused
                    #later.
                    pythonExceptionStore[e.__class__.__name__] = e.__class__
                newTb = filteredTb(sys.exc_info()[2])
                oldTb = getattr(e, 'oldTraceback', [])
                self.callbackData.result = {
                    'type': 'exception',
                    'value': {
                        'message': e.message,
                        'name': getattr(e, 'name', e.__class__.__name__),
                        # store the combined tracebacks on the JS object for
                        # later reuse
                        'oldTraceback': newTb + oldTb
                    },
                }
                return
        self.callbackData.result = {
            'type': 'result',
            'value': result,
        }

def filteredTb(tb):
    newTb = traceback.extract_tb(tb)
    return [item for item in newTb if item[0] not in JSEVAL_MASKED_FILES]

def excepthook(type, value, tb):
    if not hasattr(value, "oldTraceback"):
        #the default excepthook
        sys.__excepthook__(type, value, tb)
        return

    print >> sys.stderr, "Traceback (most recent call last):"
    #the last 2 traceback items are just JSEvaluator internals; they
    #shouldn't be frustrating the debugging process.
    print >> sys.stderr, ''.join(traceback.format_list(filteredTb(tb))),
    print >> sys.stderr, ''.join(traceback.format_list(value.oldTraceback)),
    print >> sys.stderr, "%s: %s" % (type.__name__, value)

class JSEvaluator(object):
    """JSEvaluator is an object that helps you interacting with JS code from
    Python. It allows you to modify the JS global scope in a dict-like fashion
    through its ``global_`` property, and to evaluate JavaScript code via its
    ``eval`` method, e.g.:

    ``evaluator.eval("3 + 2")``

    You can also call JS functions by accessing them in the dict-
    like way, e.g.:

    ``evaluator.global_["Math"]["ceil"](4.56)``

    Or use the ``.new()`` on a value to make an instance as with the
    JS ``new`` keyword:

    ``evaluator.global_["Date"].new()``

    For more examples, see the examples file and the tests for this module.

    Note that JS objects are proxied into Python objects, i.e. changes to the
    Python representation will change the JS value in the interpreter. The other
    way around (Python objects in JavaScript), this is not the case. Instead,
    copies are passed in.

    """
    # Error.prototype.stack parser
    JS_STACK_RE = re.compile(r'(?P<func>[^@]*)@(?P<path>[^:]*):(?P<line>[0-9]*)')
    # For quick access
    JSError = pyproxies.JSError

    def __init__(self):
        sys.excepthook = excepthook
        self.engine = QtQml.QJSEngine()
        self.engine.evaluate(INIT_SCRIPT, INIT_SCRIPT_NAME, 1)

        # provides access to the global object
        self.global_ = pyproxies.JSObject(self, self.engine.globalObject())
        cbObj = self.global_._callback
        self.callback = Callback(cbObj)
        cbObj.pyCallback = self.engine.newQObject(self.callback).property('callback')

        self.counter = itertools.count(start=1)
        self.codeStore = {}

        try:
            self.engine.installExtensions(QtQml.QJSEngine.ConsoleExtension)
        except AttributeError:
            warnings.warn("Remove this check after using only Qt >= 5.6")
            import logging
            self.global_['console'] = {
                'log': logging.debug,
                'error': logging.critical,
            }

    def eval(self, code, path=None, line=1):
        """Execute ``code`` from file ``path`` starting at line ``line`` in the
        interpreter. Note that if the last item evaluated is an error, that
        error will be raised as an exception, instead of returned as a value.
        This is a QtQml limitation. The same is true for function calls.

        """
        if not path:
            path = '<JS string %s>' % next(self.counter)
        self.codeStore[path] = code
        result = self.engine.evaluate(code, path, line)
        return self.throwIfNecessary(self.toPy(result))

    def throwIfNecessary(self, value):
        if isinstance(value, BaseException):
            raise value
        return value

    def toPy(self, value, instance=None):
        """Converts a QJSValue to a Python equivalent, often (but not always) by
        making a proxy object.

        """
        if value.isNull() or value.isUndefined():
            return None
        elif value.isNumber():
            return self.buildNumber(value)
        elif value.isBool() or value.isString():
            return value.toVariant()
        elif value.isDate():
            return value.toDateTime().toPyDateTime()
        elif value.isRegExp():
            return self.buildRegExp(value)
        elif value.isArray():
            return pyproxies.JSArray(self, value)
        elif value.isError():
            return self.buildPyError(value)
        else:
            return pyproxies.JSObject(self, value, instance)

    def buildNumber(self, value):
        """Return a Python int if an integer number, otherwise a Python float"""

        num = value.toNumber()
        try:
            asInt = int(num)
        except (ValueError, OverflowError):
            return num
        return asInt if num == asInt else num

    def buildRegExp(self, value):
        """Converts a JS regexp to its Python equivalent compiled re-object."""
        regexp = value.toVariant()
        caseInsensitive = regexp.caseSensitivity() == QtCore.Qt.CaseInsensitive
        flags = re.IGNORECASE if caseInsensitive else 0
        return re.compile(regexp.pattern(), flags)

    def buildPyError(self, value):
        """If the exception name was encountered earlier, this module will
        re-use the then-recorded exception class. Otherwise, JSError is used.

        The JS traceback is stored on the exception object, for use by the
        exception handler hook.

        """
        name = value.property('name').toString()
        message = value.property('message').toString()
        # get the relevant part of the stacktrace, i.e. the part before we
        # dive into the init script, excluding sometimes the first item (which
        # is then the wrapper function inside the init script)
        stack = value.property('stack').toString().split('\n')
        if INIT_SCRIPT_NAME in stack[0]:
            stack = stack[1:]
        stack = itertools.takewhile(lambda x: INIT_SCRIPT_NAME not in x, stack)
        stackInfo = []
        for item in stack:
            # convert JS stack line into its Python list equivalent
            m = self.JS_STACK_RE.match(item)
            path = m.group('path')
            func = m.group('func')
            if func == '%entry':
                func = '<global scope>'
            line = int(m.group('line'))
            code = self.codeStore[path].split('\n')[line - 1]
            stackInfo.insert(0, [path, line, func, code])
        try:
            err = pythonExceptionStore[name](message)
        except KeyError:
            err = pyproxies.JSError(name, message)
        oldTraceback = self.toPy(value.property('oldTraceback')) or []
        err.oldTraceback = stackInfo + list(oldTraceback)
        return err

    def toJS(self, value):
        """Convert a Python value to its corresponding QJSValue. No proxy
        objects here, as QtQml's API does not provide enough flexibility to have
        those.

        """
        if isinstance(value, pyproxies.JSValue):
            # proxy objects
            return value.val
        if value is None:
            # null
            value = QtQml.QJSValue.NullValue
        with utils.suppress(TypeError):
            # numbers, strings, bools
            return QtQml.QJSValue(value)
        if callable(value):
            # functions
            return self.wrapFunction(value)
        if isinstance(value, collections.Sequence):
            # array
            obj = pyproxies.JSArray(self, self.engine.newArray())
            obj.extend(value)
            return obj.val
        if isinstance(value, collections.Mapping):
            # object
            obj = pyproxies.JSObject(self, self.engine.newObject())
            obj.update(value)
            return obj.val
        if isinstance(value, datetime.datetime):
            # datetime
            return self.engine.evaluate('new Date("%s")' % value.isoformat())
        raise ValueError("Cannot convert object of type '%s' to a JS value" % type(value))

    def wrapFunction(self, func):
        # use the Python id, which is 1. unique and 2. the same if we get this
        # particular function another time. Both qualities are useful.
        identifier = str(id(func))
        self.callback.funcs[identifier] = func
        return self.eval('''_callback.buildJSWrapper(%r)''' % identifier).val
