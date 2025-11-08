; Функции
(function_definition
  name: (identifier) @function.name
  parameters: (parameters) @function.params
  body: (block) @function.body) @function.def

; Класове
(class_definition
  name: (identifier) @class.name
  superclasses: (argument_list)? @class.bases
  body: (block) @class.body) @class.def

; Методи (функции вътре в клас)
(class_definition
  body: (block
    (function_definition
      name: (identifier) @method.name
      parameters: (parameters) @method.params
      body: (block) @method.body))) @method.def

; Импорти
(import_statement
  name: (dotted_name) @import.module) @import

(import_from_statement
  module_name: (dotted_name) @import.from
  name: (dotted_name) @import.name) @import

; Декоратори
(decorator
  (identifier) @decorator.name) @decorator

; Docstrings
(expression_statement
  (string) @docstring) @doc
