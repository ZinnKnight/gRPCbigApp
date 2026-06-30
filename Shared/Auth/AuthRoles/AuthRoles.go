package AuthRoles

type Plan string

// в дальнейшем сделаю возможность динамической смены ролей, но пока так

const (
	Free  Plan = "free"
	Pro   Plan = "pro"
	Admin Plan = "admin"
)

func (p Plan) String() string {
	return string(p)
}

func All() []Plan {
	return []Plan{Free, Pro, Admin}
}
