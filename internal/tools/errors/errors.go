package errors

const (
	BAD_BODY                      = "переданы невалидные данные"
	CONFLICT_USER                 = "не удалось добавить пользователя, невалидные данные"
	INTERNAL_SERVER_ERROR         = "ошибка сервера"
	NOT_FOUND_USER_BY_NICK        = "Can't find user by nickname: "
	NOT_FOUND_POST_AUTHOR_BY_NICK = "Can't find post author by nickname: "
	NOT_FOUND_THREAD              = "can't find thread by slug_or_id: "
	EMAIL_ALREADY_IN_USE          = "This email is already registered by user: "
	NO_THREAD                     = "can't find thread by slug_or_id: "
	NO_PARENT_POST                = "Parent post was created in another thread"
	NO_POST                       = "can't find post by id: "
	NO_THREAD_FORUM               = "Can't find thread forum by slug: "
	UNKNOWN_SORT_TYPE             = "unknown sort type"
)
