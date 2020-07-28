package local

// listUsers builds a map of users to their "groups".
func (a *AuthProvider) listUsers() ([]*User, error) {
	file, err := a.getPasswdFile()
	if err != nil {
		return nil, err
	}

	return getAllUsersFromBuffer(file)
}

// getUser retrieves a user, their groups, and their password hash
// from the local file.
func (a *AuthProvider) getUser(username string) (*User, error) {
	file, err := a.getPasswdFile()
	if err != nil {
		return nil, err
	}
	return getUserFromBuffer(file, username)
}

// createUser adds a new user to the passwd file. If it already exists an error
// is returned.
func (a *AuthProvider) createUser(user *User) error {
	if err := a.secrets.Lock(); err != nil {
		return err
	}
	defer a.secrets.Release()
	file, err := a.getPasswdFile()
	if err != nil {
		return err
	}
	// addUserToBuffer returns an error if it finds a matching user in the file
	// already
	newFile, err := addUserToBuffer(file, user)
	if err != nil {
		return err
	}
	return a.updatePasswdFile(newFile)
}

func (a *AuthProvider) updateUser(user *User) error {
	if err := a.secrets.Lock(); err != nil {
		return err
	}
	defer a.secrets.Release()
	file, err := a.getPasswdFile()
	if err != nil {
		return err
	}
	newFile, err := updateUserInBuffer(file, user)
	if err != nil {
		return err
	}
	return a.updatePasswdFile(newFile)
}

func (a *AuthProvider) deleteUser(username string) error {
	if err := a.secrets.Lock(); err != nil {
		return err
	}
	defer a.secrets.Release()
	file, err := a.getPasswdFile()
	if err != nil {
		return err
	}
	newFile, err := deleteUserInBuffer(file, username)
	if err != nil {
		return err
	}
	return a.updatePasswdFile(newFile)
}
