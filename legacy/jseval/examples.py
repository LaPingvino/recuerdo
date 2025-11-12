"""Some examples, including seeing mixed Python/JS exceptions in action"""

from PyQt5 import QtWidgets
import __init__ as jseval

app = QtWidgets.QApplication([])

js = jseval.JSEvaluator()
js.eval('var z = {x: 10, y: function x(a) {return this.x + a[0].b; }}')
print(js.global_['z']['y']([{'b': 4}]))

js.global_['pow3'] = lambda x: x * x * x
print(js.global_['pow3'](2))
js.global_['twice'] = lambda x: x + x
print(js.global_['twice']('test'))

def pyFunc3():
    raise ValueError("Hello World!")

def pyFunc2():
    pyFunc3()

def pyFunc():
    js.global_.jsFunc2()

js.global_["pyFunc"] = pyFunc
js.global_["pyFunc2"] = pyFunc2

program = """
    function jsFunc2() {
        pyFunc2();
    }
    function jsFunc() {
        pyFunc();
    }
    jsFunc();
"""
js.eval(program)
