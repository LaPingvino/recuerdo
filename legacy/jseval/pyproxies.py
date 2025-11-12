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

import collections
from PyQt5 import QtQml
import copy

class JSValue(object):
    """"Behaviour shared between arrays and objects - useless on its own."""

    NotFound = KeyError

    def __init__(self, evaluator, jsValue):
        # __dict__ to circumvent the custom __setattr__ implementation
        self.__dict__['js'] = evaluator
        self.__dict__['val'] = jsValue

    def _get(self, key):
        return self.js.toPy(self.val.property(key), self.val)

    def __setitem__(self, key, value):
        self.val.setProperty(key, self.js.toJS(value))

    def _possiblyRecursive(self, property):
        return isinstance(property, self.__class__)

class JSArray(JSValue, collections.MutableSequence):
    """A proxy object that makes it possible to threat an Array QJSValue like a
    Python list

    """
    NotFound = IndexError

    def __iter__(self):
        for i in range(len(self)):
            yield self[i]

    def insert(self, idx, value):
        self._splice([idx, 0, self.js.toJS(value)])

    def _splice(self, args):
        args = map(QtQml.QJSValue, args)
        self.val.property('splice').callWithInstance(self.val, args)

    def _normalize(self, idx):
        if idx < 0:
            idx = len(self) + idx
        if not 0 <= idx < len(self):
            raise self.NotFound(idx)
        return idx

    def __getitem__(self, idx):
        return self._get(self._normalize(idx))

    def __delitem__(self, idx):
        self._splice([self._normalize(idx), 1])

    def __len__(self):
        return self.val.property('length').toInt()

    def __repr__(self):
        items = ('[...]' if self._possiblyRecursive(item) else repr(item) for item in self)
        return '<%s[%s]>' % (self.__class__.__name__, ', '.join(items))

    def __copy__(self):
        return list(self)

    def __deepcopy__(self, memo):
        return [copy.deepcopy(item, memo) for item in self]

    def __eq__(self, other):
        return len(self) == len(other) and all(a == b for a, b in zip(self, other))

class KeysAsAttributesMixin(object):
    """Makes it possible to access the keys of a dictionary-like as properties
    (assuming there is no actual property of the same name that takes
    precendence)

    """
    def __getattr__(self, attr):
        try:
            return super(KeysAsAttributesMixin, self).__getattr__(attr)
        except AttributeError:
            try:
                return self[attr]
            except KeyError, e:
                raise AttributeError(e)

    def __setattr__(self, attr, value):
        self[attr] = value

    def __delattr__(self, attr):
        try:
            del self[attr]
        except KeyError, e:
            raise AttributeError(e)

class JSObject(JSValue, KeysAsAttributesMixin, collections.MutableMapping):
    """A proxy object that makes it possible to threat an Object QJSValue like a
    Python dictionary (and also a normal Python object).

    Note that JS functions are also objects, and are thus also wrapped in this
    class. To cater to them, ``__call__`` and ``new()`` are defined.

    """
    def __init__(self, evaluator, jsValue, instance=None):
        super(JSObject, self).__init__(evaluator, jsValue)

        # __dict__ to circumvent the custom __setattr__ implementation
        self.__dict__['instance'] = instance

    def __eq__(self, other):
        try:
            return len(self) == len(other) and all(v == other[k] for k, v in self.iteritems())
        except TypeError:
            return False

    def __getitem__(self, key):
        if self.val.hasProperty(key):
            return self._get(key)
        raise self.NotFound(key)

    def __delitem__(self, key):
        if not self.val.hasOwnProperty(key):
            raise self.NotFound(key)
        self.val.deleteProperty(key)

    def __iter__(self):
        iterator = QtQml.QJSValueIterator(self.val)
        while iterator.next():
            yield iterator.name()

    def __len__(self):
        return sum(1 for _ in self)

    def _handleCall(self, callFunc, args, kwargs):
        assert self.val.isCallable()
        moddedArgs = [self.js.toJS(a) for a in args]
        if kwargs:
            # kwargs is added as an object as the last argument
            moddedArgs += [self.js.toJS(kwargs)]
        return self.js.throwIfNecessary(self.js.toPy(callFunc(moddedArgs)))

    def new(self, *args, **kwargs):
        return self._handleCall(self.val.callAsConstructor, args, kwargs)

    def __call__(self, *args, **kwargs):
        thisValue = self.instance or self.js.global_.val
        callFunc = lambda args: self.val.callWithInstance(thisValue, args)
        return self._handleCall(callFunc, args, kwargs)

    def __repr__(self):
        pairs = ((key, self[key]) for key in self)
        filtered = ((k, '{...}' if self._possiblyRecursive(v) else v) for k, v in pairs)
        strings = ("%s: %s" % pair for pair in filtered)
        return "<%s{%s}>" % (self.__class__.__name__, ', '.join(strings))

    def __copy__(self):
        return JSObjectCopy(self)

    copy = __copy__

    def __deepcopy__(self, memo):
        return JSObjectCopy((key, copy.deepcopy(value, memo)) for key, value in self.iteritems())

class JSObjectCopy(KeysAsAttributesMixin, dict):
    pass

class JSError(Exception):
    """Represents a JS error. No proxying, but as error objects are not normally
    modified after their initialization that's not a big loss. The advantage
    (nice Python integration) is worth it.

    """
    def __init__(self, name, message):
        self.name = name
        self.message = message

    def __str__(self):
        return "%s: %s" % (self.name, self.message)
