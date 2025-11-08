; Функции
(function_declaration
  name: (identifier) @function.name
  parameters: (parameter_list) @function.params
  result: (_)? @function.return
  body: (block) @function.body) @function.def

; Методи
(method_declaration
  receiver: (parameter_list) @method.receiver
  name: (field_identifier) @method.name
  parameters: (parameter_list) @method.params
  result: (_)? @method.return
  body: (block) @method.body) @method.def

; Структури
(type_declaration
  (type_spec
    name: (type_identifier) @struct.name
    type: (struct_type) @struct.body)) @struct.def

; Интерфейси
(type_declaration
  (type_spec
    name: (type_identifier) @interface.name
    type: (interface_type) @interface.body)) @interface.def

; Импорти
(import_declaration
  (import_spec
    path: (interpreted_string_literal) @import.path)) @import

; Коментари
(comment) @comment

; Константи
(const_declaration
  (const_spec
    name: (identifier) @const.name
    value: (_) @const.value)) @const.def
