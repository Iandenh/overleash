#include <php.h>
#include <Zend/zend_API.h>
#include <Zend/zend_hash.h>
#include <Zend/zend_types.h>
#include <zend_exceptions.h>
#include <stddef.h>

#include "_cgo_export.h"
#include "overleash.h"
#include "overleash_arginfo.h"

PHP_FUNCTION(Iandenh_Overleash_isEnabled) {
    zend_string *feature_name = NULL;
    zend_array *context = NULL;

    ZEND_PARSE_PARAMETERS_START(1, 2)
        Z_PARAM_STR(feature_name)
        Z_PARAM_OPTIONAL
        Z_PARAM_ARRAY_HT(context)
    ZEND_PARSE_PARAMETERS_END();

    struct go_is_enabled_return ret = go_is_enabled(feature_name, context);

    if (ret.r1 != NULL) {
        zend_throw_exception(NULL, ret.r1, 0);
        free(ret.r1);
        RETURN_THROWS();
    }

    RETURN_BOOL(ret.r0);
}

zend_module_entry overleash_module_entry = {
    STANDARD_MODULE_HEADER,
    "frankenphp-overleash",
    ext_functions, /* Functions */
    NULL, /* MINIT */
    NULL, /* MSHUTDOWN */
    NULL, /* RINIT */
    NULL, /* RSHUTDOWN */
    NULL, /* MINFO */
    "0.1.0", /* Version */
    STANDARD_MODULE_PROPERTIES
};