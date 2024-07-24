#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

/**
 * Instantiates a new engine. Returns a pointer to the engine.
 *
 * # Safety
 *
 * The caller is responsible for freeing the allocated memory. This can be done by calling
 * `free_engine` and passing in the pointer returned by this method. Failure to do so will result in a leak.
 */
void *new_engine(void);

/**
 * Frees the memory allocated for the engine.
 *
 * # Safety
 *
 * The caller is responsible for ensuring the argument is a valid pointer.
 * Null pointers will result in a no-op, but any invalid pointers will result in undefined behavior.
 * These pointers should not be dropped for the lifetime of this function call.
 *
 * This function must be called correctly in order to deallocate the memory allocated for the engine in
 * the `new_engine` function. Failure to do so will result in a leak.
 */
void free_engine(void *engine_ptr);

/**
 * Takes a JSON string representing a set of toggles. Returns a JSON encoded response object
 * specifying whether the update was successful or not. The caller is responsible
 * for freeing this response object.
 *
 * # Safety
 *
 * The caller is responsible for ensuring all arguments are valid pointers.
 * Null pointers will result in an error message being returned to the caller,
 * but any invalid pointers will result in undefined behavior.
 * These pointers should not be dropped for the lifetime of this function call.
 */
const char *take_state(void *engine_ptr, const char *json_ptr);

/**
 * Checks if a toggle is enabled for a given context. Returns a JSON encoded response of type `EnabledResponse`.
 *
 * # Safety
 *
 * The caller is responsible for ensuring all arguments are valid pointers.
 * Null pointers will result in an error message being returned to the caller,
 * but any invalid pointers will result in undefined behavior.
 * These pointers should not be dropped for the lifetime of this function call.
 *
 * The caller is responsible for freeing the allocated memory. This can be done by calling
 * `free_response` and passing in the pointer returned by this method. Failure to do so will result in a leak.
 */
const char *check_enabled(void *engine_ptr,
                          const char *toggle_name_ptr,
                          const char *context_ptr);

const char *resolve_all(void *engine_ptr, const char *context_ptr);

const char *resolve(void *engine_ptr, const char *toggle_name_ptr, const char *context_ptr);

/**
 * Checks the toggle variant for a given context. Returns a JSON encoded response of type `VariantResponse`.
 *
 * # Safety
 *
 * The caller is responsible for ensuring all arguments are valid pointers.
 * Null pointers will result in an error message being returned to the caller,
 * but any invalid pointers will result in undefined behavior.
 * These pointers should not be dropped for the lifetime of this function call.
 *
 * The caller is responsible for freeing the allocated memory. This can be done by calling
 * `free_response` and passing in the pointer returned by this method. Failure to do so will result in a leak.
 */
const char *check_variant(void *engine_ptr,
                          const char *toggle_name_ptr,
                          const char *context_ptr);

/**
 * Returns a JSON encoded response with a list of strings representing the built-in strategies Yggdrasil supports.
 *
 * # Safety
 *
 * The caller is responsible for freeing the allocated memory. This can be done by calling
 * `free_response` and passing in the pointer returned by this method. Failure to do so will result in a leak.
 */
const char *built_in_strategies(void);

/**
 * Frees the memory allocated for a response message created by `check_enabled` or `check_variant`.
 *
 * # Safety
 *
 * The caller is responsible for ensuring all arguments are valid pointers.
 * Null pointers will result in an error message being returned to the caller,
 * but any invalid pointers will result in undefined behavior.
 * These pointers should not be dropped for the lifetime of this function call.
 *
 * This function must be called correctly in order to deallocate the memory allocated for the response in
 * the `check_enabled`, `check_variant`, `count_toggle`, `count_variant` and `get_metrics` functions.
 * Failure to do so will result in a leak.
 */
void free_response(char *response_ptr);

/**
 * Marks a toggle as being counted for purposes of metrics. This function needs to be paired with a call
 * to `get_metrics` at a later point in time to retrieve the metrics.
 *
 * # Safety
 *
 * The caller is responsible for ensuring all arguments (except the last one, `enabled`) are valid pointers.
 * Null pointers will result in an error message being returned to the caller,
 * but any invalid pointers will result in undefined behavior.
 * These pointers should not be dropped for the lifetime of this function call.
 *
 * The caller is responsible for freeing the allocated memory. This can be done by calling
 * `free_response` and passing in the pointer returned by this method. Failure to do so will result in a leak.
 */
const char *count_toggle(void *engine_ptr,
                         const char *toggle_name_ptr,
                         bool enabled);

/**
 * Marks a variant as being counted for purposes of metrics. This function needs to be paired with a call
 * to `get_metrics` at a later point in time to retrieve the metrics.
 *
 * # Safety
 *
 * The caller is responsible for ensuring all arguments are valid pointers.
 * Null pointers will result in an error message being returned to the caller,
 * but any invalid pointers will result in undefined behavior.
 * These pointers should not be dropped for the lifetime of this function call.
 *
 * The caller is responsible for freeing the allocated memory. This can be done by calling
 * `free_response` and passing in the pointer returned by this method. Failure to do so will result in a leak.
 */
const char *count_variant(void *engine_ptr,
                          const char *toggle_name_ptr,
                          const char *variant_name_ptr);

/**
 * Returns a JSON encoded string representing the number of times each toggle and variant has been
 * counted since the last time this function was called.
 *
 * # Safety
 *
 * The caller is responsible for ensuring all arguments are valid pointers.
 * Null pointers will result in an error message being returned to the caller,
 * but any invalid pointers will result in undefined behavior.
 * These pointers should not be dropped for the lifetime of this function call.
 *
 * The caller is responsible for freeing the allocated memory, in case the response is not null. This can be done by calling
 * `free_response` and passing in the pointer returned by this method. Failure to do so will result in a leak.
 */
char *get_metrics(void *engine_ptr);

/**
 * Lets you know whether impression events are enabled for this toggle or not.
 * Returns a JSON encoded response of type `Response`.
 *
 * # Safety
 *
 * The caller is responsible for ensuring the engine_ptr is a valid pointer to an unleash engine.
 * An invalid pointer to unleash engine will result in undefined behaviour.
 */
char *should_emit_impression_event(void *engine_ptr, const char *toggle_name_ptr);
